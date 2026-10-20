"""Collect the pinned LiveKit binary's module inventory and license texts."""
import argparse, concurrent.futures, json, pathlib, subprocess, zipfile, time
p=argparse.ArgumentParser()
p.add_argument('--go',required=True);p.add_argument('--cache',required=True)
a=p.parse_args();root=pathlib.Path(__file__).resolve().parents[1]
pin=json.loads((root/'engine.lock.json').read_text(encoding='utf-8'))
archive=pathlib.Path(a.cache)/pin['asset']
engine=pathlib.Path(a.cache)/'livekit-license-inventory.exe'
with zipfile.ZipFile(archive) as z: engine.write_bytes(z.read('livekit-server.exe'))
info=subprocess.check_output([a.go,'version','-m',str(engine)],text=True)
modules=[]
for line in info.splitlines():
    parts=line.strip().split()
    if len(parts)>=3 and parts[0] in ('dep','mod'):
        if parts[2]=='(devel)': continue
        modules.append((parts[1],parts[2]))
def fetch(m):
    name,version=m
    for attempt in range(3):
        result=subprocess.run([a.go,'mod','download','-json',name+'@'+version],text=True,capture_output=True,timeout=180)
        if result.returncode==0: break
        if attempt==2: raise RuntimeError(name+': '+result.stdout+result.stderr)
        time.sleep(2)
    text=result.stdout
    metadata=json.loads(text);directory=pathlib.Path(metadata['Dir'])
    files=[f for f in directory.iterdir() if f.is_file() and f.name.upper().startswith(('LICENSE','LICENCE','COPYING','NOTICE'))]
    if not files: raise RuntimeError('Missing license: '+name)
    print('Collected '+name+' '+version,flush=True)
    return {'name':name,'version':version,'sum':metadata.get('Sum')},'\n===== '+name+' '+version+' =====\n'+'\n'.join(f.read_text(encoding='utf-8',errors='replace') for f in files)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    results=list(pool.map(fetch,modules))
(root/'docs/LIVEKIT-DEPENDENCIES.json').write_text(json.dumps([r[0] for r in results],indent=2)+'\n',encoding='utf-8')
(root/'docs/LIVEKIT-DEPENDENCY-LICENSES.txt').write_text('Licenses of modules recorded in the pinned LiveKit Windows binary.\n'+'\n'.join(r[1] for r in results),encoding='utf-8')
engine.unlink()
print('Recorded LiveKit licenses:',len(results))
