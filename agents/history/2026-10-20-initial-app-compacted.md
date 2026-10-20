# Initial application — implementation and release preparation

User requested My Maps and My Meet as native Windows apps in One Workspace with shared design, SSO, standalone use, installation and public org releases. Free/open source foundations only; no GitHub Actions. UI uses AUI compiler with an adapter derived from Apache-2.0 My Tasks; shared apphost pinned to 0578d36ce297. Local Go/TypeScript/Vite builds passed. Isolated NTFS preview at C:/tmp/maps-meet-preview-20261020 only, no owner data used. Meet preview opens a real LiveKit room, publishes synthetic audio/video, admits a second account and displays meeting chat. Maps street tiles load; Photon search/results also work; route interaction has not received separate end-to-end acceptance. These checks are not physical-device, public-network or capacity acceptance.

Source is I:/Projects/my-meet; suite integration is I:/Projects/one-workspace-maps-meet branch codex/maps-meet. Build material lives in I:/Projects/tooling/maps-meet-20261020. No release or owner install yet. Preserve explicit README preview boundaries, provider policies, engine licence inventory and private-data boundaries. Continue by completing package/build receipts, org publication, combined Workspace release and normal backed-up owner update. Reconcile newer primary Workspace commits before final tag. Clean only named synthetic C fixture and owned child processes afterward.

Preview1 upload was held as an unpublished draft after source review found camera mute/unmute DOM reattachment and declined-lobby feedback issues. Preview2 corrects both; do not rewrite the preview1 source tag. The HTTP tunnel still needs a separate media network path for remote calling.

Integrated navigation follow-up: explicit shared-picker trigger suppresses the floating launcher, and compact labelled icon controls keep the managed mobile header within the viewport. Meet also binds its signalling grants to the current authenticated app session. Maps2 and Meet3 supersede the initial previews in Workspace54; do not rewrite published assets/tags.

Final integrated minimum is Workspace54, incorporating the concurrent Clips4 restart repair. Initial Maps1 and Meet2 releases are public; Maps2/Meet3 are the final integrated candidates. Owner installation remains pending.

## Delivery checkpoint

Meet3 sourceb39d099 is public with16 verified release assets including its
managed-update manifest. The final packaged preview passed two synthetic Chrome
clients: host admission, shared chat, media reception and camera off/on remote
decode with session-bound signalling. Normal owner53-to54 update installed it
alongside Maps2;24 apps and23 services healthy, identity/storage retained.
The public route returns the expected login redirect. Managed browser SSO was
not rechecked because Chrome control could not load its request-header policy.
Synthetic preview data was removed from C. Public media/NAT acceptance is still
pending the owner's router-forwarding choice; no paid relay is enabled.

Combined55 supersedes the held54 draft to fix inherited release-date ordering.
Drive3 and suite55 publication remain pending slow large-file uploads. Do not
claim full Google parity, internet-call acceptance or final suite publication.

Owner manager 55 update is now complete: all 24 app packages/versions and identity/
storage fields preserved, 23 service entries healthy, newest rollback verified.
Meet 3 remains installed unchanged. Full suite 55/Drive 3 publication is still
pending GitHub HTTP 408 large-transfer retries; standalone Meet 3 is public.
