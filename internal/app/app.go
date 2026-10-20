package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oneworkspacestudio/my-meet/internal/engine"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type App struct {
	H              *apphost.Host
	engine         string
	port, udp, tcp int
	key, secret    string
	client         *http.Client
}
type Meeting struct {
	Code   string `json:"code"`
	Title  string `json:"title"`
	Owner  string `json:"owner"`
	Starts string `json:"starts"`
	Locked bool   `json:"locked"`
	Ended  bool   `json:"ended"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

const schema = `CREATE TABLE meetings(code TEXT PRIMARY KEY,title TEXT NOT NULL,owner TEXT NOT NULL,starts TEXT NOT NULL DEFAULT '',locked INTEGER NOT NULL DEFAULT 0,ended INTEGER NOT NULL DEFAULT 0,created TEXT NOT NULL);
 CREATE TABLE members(room TEXT NOT NULL REFERENCES meetings(code),user TEXT NOT NULL,name TEXT NOT NULL,status TEXT NOT NULL,PRIMARY KEY(room,user));
 CREATE TABLE settings(key TEXT PRIMARY KEY,value TEXT NOT NULL);
 CREATE TABLE messages(id INTEGER PRIMARY KEY,room TEXT NOT NULL,owner TEXT NOT NULL,name TEXT NOT NULL,body TEXT NOT NULL,created TEXT NOT NULL);
 CREATE INDEX message_room ON messages(room,id);`

func New(h *apphost.Host) (*App, error) {
	if e := h.Migrate(1, schema); e != nil {
		return nil, e
	}
	exe, e := engine.Prepare(h.Config.Data)
	if e != nil {
		return nil, e
	}
	_, port, _ := net.SplitHostPort(h.Config.Listen)
	n, _ := strconv.Atoi(port)
	if n > 64000 {
		return nil, errors.New("Meet listener must use a port below 64001")
	}
	a := &App{H: h, engine: exe, port: n + 1000, udp: n + 1001, tcp: n + 1002, client: &http.Client{Timeout: 10 * time.Second}}
	for k, d := range map[string]*string{"engine-key": &a.key, "engine-secret": &a.secret} {
		if h.DB.QueryRow("SELECT value FROM settings WHERE key=?", k).Scan(d) != nil {
			*d = apphost.ID() + apphost.ID()
			if _, e = h.DB.Exec("INSERT INTO settings VALUES(?,?)", k, *d); e != nil {
				return nil, e
			}
		}
	}
	h.Mux.HandleFunc("GET /api/config", h.Require(a.config))
	h.Mux.HandleFunc("GET /api/meetings", h.Require(a.list))
	h.Mux.HandleFunc("POST /api/meetings", h.Require(a.create))
	h.Mux.HandleFunc("GET /api/meetings/{code}", h.Require(a.detail))
	h.Mux.HandleFunc("POST /api/meetings/{code}/join", h.Require(a.join))
	h.Mux.HandleFunc("POST /api/meetings/{code}/host", h.Require(a.host))
	h.Mux.HandleFunc("GET /api/meetings/{code}/messages", h.Require(a.messages))
	h.Mux.HandleFunc("POST /api/meetings/{code}/messages", h.Require(a.message))
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", a.port))
	proxy := httputil.NewSingleHostReverseProxy(target)
	original := proxy.Director
	proxy.Director = func(r *http.Request) {
		original(r)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/signal")
		r.Header.Del("Cookie")
		r.Header.Del("Authorization")
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		apphost.Problem(w, 503, "Meeting engine is starting. Please retry in a moment.")
	}
	// Expose only the LiveKit client transport. Admin Twirp APIs stay on loopback.
	h.Mux.HandleFunc("GET /signal/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/signal")
		if !(path == "/rtc" || path == "/rtc/validate" || path == "/rtc/v1" || path == "/rtc/v1/validate") {
			http.NotFound(w, r)
			return
		}
		if !a.permittedToken(r.URL.Query().Get("access_token")) {
			apphost.Problem(w, 403, "Rejoin the meeting to refresh your permission")
			return
		}
		origin := r.Header.Get("Origin")
		if origin != "" {
			u, e := url.Parse(origin)
			if e != nil || u.Host != r.Host {
				apphost.Problem(w, 403, "Open Meet from its configured address")
				return
			}
		}
		proxy.ServeHTTP(w, r)
	})
	return a, nil
}
func (a *App) permittedToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(a.secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	sig, e := base64.RawURLEncoding.DecodeString(parts[2])
	if e != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return false
	}
	body, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return false
	}
	var c struct {
		Iss   string `json:"iss"`
		Sub   string `json:"sub"`
		Exp   int64  `json:"exp"`
		Video struct {
			Room string `json:"room"`
			Join bool   `json:"roomJoin"`
		} `json:"video"`
	}
	if json.Unmarshal(body, &c) != nil || c.Iss != a.key || c.Exp <= time.Now().Unix() || !c.Video.Join {
		return false
	}
	m, e := a.get(c.Video.Room, c.Sub)
	return e == nil && !m.Ended && m.Status == "admitted"
}
func (a *App) Run(ctx context.Context) error {
	config := map[string]any{"port": a.port, "bind_addresses": []string{"127.0.0.1"}, "keys": map[string]string{a.key: a.secret}, "rtc": map[string]any{"udp_port": a.udp, "tcp_port": a.tcp, "use_external_ip": os.Getenv("MYMEET_EXTERNAL_IP") == "true", "use_ice_lite": false}, "room": map[string]any{"auto_create": true, "max_participants": 32, "empty_timeout": 120}, "logging": map[string]any{"level": "error"}}
	data, _ := json.Marshal(config)
	return apphost.Service(ctx, a.engine, a.H.Config.Data, nil, []string{"LIVEKIT_CONFIG=" + string(data)}, io.Discard)
}
func (a *App) token(subject, name string, grants map[string]any, ttl time.Duration) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims, _ := json.Marshal(map[string]any{"iss": a.key, "sub": subject, "name": name, "nbf": time.Now().Add(-10 * time.Second).Unix(), "exp": time.Now().Add(ttl).Unix(), "video": grants})
	body := header + "." + base64.RawURLEncoding.EncodeToString(claims)
	mac := hmac.New(sha256.New, []byte(a.secret))
	mac.Write([]byte(body))
	return body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (a *App) rpc(ctx context.Context, room, method string, input any) error {
	b, _ := json.Marshal(input)
	req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://127.0.0.1:%d/twirp/livekit.RoomService/%s", a.port, method), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+a.token("manager", "My Meet", map[string]any{"roomAdmin": true, "roomCreate": true, "roomList": true, "room": room}, time.Minute))
	req.Header.Set("Content-Type", "application/json")
	resp, e := a.client.Do(req)
	if e != nil {
		return errors.New("meeting engine is unavailable")
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == 404 {
		return nil
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("meeting engine declined the action (%d)", resp.StatusCode)
	}
	return nil
}
func (a *App) config(w http.ResponseWriter, r *http.Request) {
	resp, e := a.client.Get(fmt.Sprintf("http://127.0.0.1:%d/", a.port))
	ready := e == nil && resp.StatusCode == 200
	if resp != nil {
		resp.Body.Close()
	}
	apphost.JSON(w, 200, map[string]any{"ready": ready, "udpPort": a.udp, "tcpPort": a.tcp, "externalIP": os.Getenv("MYMEET_EXTERNAL_IP") == "true", "maxParticipants": 32, "networkNotice": "Remote calls require the media ports to reach this PC. A Cloudflare HTTP tunnel carries signalling only. For restricted networks use a separately hosted TURN relay."})
}
func (a *App) get(code, user string) (Meeting, error) {
	var m Meeting
	e := a.H.DB.QueryRow("SELECT code,title,owner,starts,locked,ended FROM meetings WHERE code=?", code).Scan(&m.Code, &m.Title, &m.Owner, &m.Starts, &m.Locked, &m.Ended)
	if e != nil {
		return m, e
	}
	m.Role = "member"
	m.Status = "new"
	if m.Owner == user {
		m.Role = "host"
		m.Status = "admitted"
	} else {
		a.H.DB.QueryRow("SELECT status FROM members WHERE room=? AND user=?", code, user).Scan(&m.Status)
	}
	return m, nil
}
func (a *App) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title  string `json:"title"`
		Starts string `json:"starts"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		in.Title = "Instant meeting"
	}
	if len(in.Title) > 200 {
		apphost.Problem(w, 400, "Use a shorter meeting title")
		return
	}
	if in.Starts != "" {
		if _, e := time.Parse(time.RFC3339, in.Starts); e != nil {
			apphost.Problem(w, 400, "Use a valid meeting date")
			return
		}
	}
	raw := apphost.ID()[:20]
	code := raw[:5] + "-" + raw[5:10] + "-" + raw[10:15] + "-" + raw[15:]
	u := apphost.Current(r).User
	_, e := a.H.DB.Exec("INSERT INTO meetings(code,title,owner,starts,created) VALUES(?,?,?,?,?)", code, in.Title, u.ID, in.Starts, time.Now().UTC().Format(time.RFC3339))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	m, _ := a.get(code, u.ID)
	apphost.JSON(w, 201, m)
}
func (a *App) list(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User.ID
	rows, e := a.H.DB.Query("SELECT code,title,owner,starts,locked,ended FROM meetings WHERE owner=? OR code IN(SELECT room FROM members WHERE user=? AND status='admitted') ORDER BY created DESC LIMIT 100", u, u)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Meeting{}
	for rows.Next() {
		var m Meeting
		if e = rows.Scan(&m.Code, &m.Title, &m.Owner, &m.Starts, &m.Locked, &m.Ended); e != nil {
			a.H.DBError(w, e)
			return
		}
		m.Role = "member"
		if m.Owner == u {
			m.Role = "host"
		}
		out = append(out, m)
	}
	apphost.JSON(w, 200, out)
}
func (a *App) detail(w http.ResponseWriter, r *http.Request) {
	m, e := a.get(r.PathValue("code"), apphost.Current(r).User.ID)
	if e != nil {
		apphost.Problem(w, 404, "Meeting not found. Check the link or code.")
		return
	}
	members := []map[string]string{}
	if m.Status == "admitted" {
		rows, e := a.H.DB.Query("SELECT user,name,status FROM members WHERE room=?", m.Code)
		if e != nil {
			a.H.DBError(w, e)
			return
		}
		for rows.Next() {
			var id, name, status string
			rows.Scan(&id, &name, &status)
			if m.Role == "host" || status == "admitted" {
				members = append(members, map[string]string{"id": id, "name": name, "status": status})
			}
		}
		rows.Close()
	}
	apphost.JSON(w, 200, map[string]any{"meeting": m, "members": members})
}
func (a *App) join(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User
	m, e := a.get(r.PathValue("code"), u.ID)
	if e != nil {
		apphost.Problem(w, 404, "Meeting not found")
		return
	}
	if m.Ended {
		apphost.Problem(w, 410, "The host ended this meeting")
		return
	}
	if m.Status == "denied" || m.Status == "removed" {
		apphost.Problem(w, 403, "The host has not allowed you to join this meeting")
		return
	}
	if m.Status != "admitted" {
		if m.Locked {
			apphost.Problem(w, 403, "This meeting is locked. Ask the host to unlock it.")
			return
		}
		_, e = a.H.DB.Exec("INSERT INTO members(room,user,name,status)VALUES(?,?,?,'waiting') ON CONFLICT(room,user) DO UPDATE SET name=excluded.name", m.Code, u.ID, u.Name)
		if e != nil {
			a.H.DBError(w, e)
			return
		}
		apphost.JSON(w, 202, map[string]string{"status": "waiting"})
		return
	}
	_, e = a.H.DB.Exec("INSERT INTO members(room,user,name,status)VALUES(?,?,?,'admitted') ON CONFLICT(room,user) DO UPDATE SET name=excluded.name", m.Code, u.ID, u.Name)
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	token := a.token(u.ID, u.Name, map[string]any{"roomJoin": true, "room": m.Code, "canPublish": true, "canSubscribe": true, "canPublishData": true, "canUpdateOwnMetadata": true}, 30*time.Second)
	apphost.JSON(w, 200, map[string]any{"status": "admitted", "token": token, "signal": a.H.Base + "/signal", "meeting": m})
}
func (a *App) host(w http.ResponseWriter, r *http.Request) {
	m, e := a.get(r.PathValue("code"), apphost.Current(r).User.ID)
	if e != nil || m.Role != "host" {
		apphost.Problem(w, 403, "Only the meeting host can do this")
		return
	}
	var in struct {
		Action string `json:"action"`
		User   string `json:"user"`
		Track  string `json:"track"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	switch in.Action {
	case "admit", "deny":
		status := "admitted"
		if in.Action == "deny" {
			status = "denied"
		}
		_, e = a.H.DB.Exec("UPDATE members SET status=? WHERE room=? AND user=? AND status='waiting'", status, m.Code, in.User)
	case "lock", "unlock":
		_, e = a.H.DB.Exec("UPDATE meetings SET locked=? WHERE code=?", in.Action == "lock", m.Code)
	case "remove":
		if in.User == m.Owner {
			apphost.Problem(w, 400, "Use Leave meeting or End for everyone")
			return
		}
		// Revoke future joins before disconnecting; existing short tokens expire in 30s.
		_, e = a.H.DB.Exec("UPDATE members SET status='removed' WHERE room=? AND user=?", m.Code, in.User)
		if e == nil {
			e = a.rpc(r.Context(), m.Code, "RemoveParticipant", map[string]string{"room": m.Code, "identity": in.User})
		}
	case "mute":
		e = a.rpc(r.Context(), m.Code, "MutePublishedTrack", map[string]any{"room": m.Code, "identity": in.User, "track_sid": in.Track, "muted": true})
	case "end":
		_, e = a.H.DB.Exec("UPDATE meetings SET ended=1,locked=1 WHERE code=?", m.Code)
		if e == nil {
			e = a.rpc(r.Context(), m.Code, "DeleteRoom", map[string]string{"room": m.Code})
		}
	default:
		apphost.Problem(w, 400, "Unknown host action")
		return
	}
	if e != nil {
		apphost.Problem(w, 503, "The host action could not finish. Please retry.")
		return
	}
	a.H.Audit(apphost.Current(r).User.ID, "meet."+in.Action, m.Code)
	apphost.JSON(w, 200, map[string]bool{"ok": true})
}
func (a *App) allowed(w http.ResponseWriter, r *http.Request) bool {
	m, e := a.get(r.PathValue("code"), apphost.Current(r).User.ID)
	if e != nil || m.Status != "admitted" || m.Ended {
		apphost.Problem(w, 403, "Join the meeting before opening its conversation")
		return false
	}
	return true
}
func (a *App) messages(w http.ResponseWriter, r *http.Request) {
	if !a.allowed(w, r) {
		return
	}
	rows, e := a.H.DB.Query("SELECT id,name,body,created FROM (SELECT * FROM messages WHERE room=? ORDER BY id DESC LIMIT 200) ORDER BY id", r.PathValue("code"))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int
		var name, body, at string
		rows.Scan(&id, &name, &body, &at)
		out = append(out, map[string]any{"id": id, "name": name, "body": body, "at": at})
	}
	apphost.JSON(w, 200, out)
}
func (a *App) message(w http.ResponseWriter, r *http.Request) {
	if !a.allowed(w, r) {
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	in.Body = strings.TrimSpace(in.Body)
	if in.Body == "" || len(in.Body) > 8000 {
		apphost.Problem(w, 400, "Messages must contain 1–8000 characters")
		return
	}
	u := apphost.Current(r).User
	var recent int
	a.H.DB.QueryRow("SELECT count(*) FROM messages WHERE owner=? AND created>?", u.ID, time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)).Scan(&recent)
	if recent >= 30 {
		apphost.Problem(w, 429, "Please slow down and try again in a moment")
		return
	}
	_, e := a.H.DB.Exec("INSERT INTO messages(room,owner,name,body,created)VALUES(?,?,?,?,?)", r.PathValue("code"), u.ID, u.Name, in.Body, time.Now().UTC().Format(time.RFC3339Nano))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 201, map[string]bool{"ok": true})
}
