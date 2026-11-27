"""Assemble the Linux plugin using Mattermost's directory-first archive layout."""
import argparse
import json
from pathlib import Path
import shutil
import tarfile
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument("source", type=Path)
parser.add_argument("output", type=Path)
opts = parser.parse_args()
source = opts.source.resolve()
manifest = json.loads((source / "plugin.json").read_text())
plugin_id = manifest["id"]
binary = "server/dist/plugin-linux-amd64"
manifest["server"]["executables"] = {"linux-amd64": binary}
with tempfile.TemporaryDirectory() as tmp:
    bundle = Path(tmp) / plugin_id
    bundle.mkdir()
    for folder in ("assets", "public", "webapp/dist"):
        if (source / folder).is_dir():
            shutil.copytree(source / folder, bundle / folder)
    (bundle / binary).parent.mkdir(parents=True)
    shutil.copy2(source / binary, bundle / binary)
    shutil.copy2(source / "LICENSE", bundle / "LICENSE")
    licenses = bundle / "assets/licenses/browser-whisper"
    licenses.mkdir(parents=True, exist_ok=True)
    for name in ("LICENSE", "THIRD_PARTY_NOTICES.md"):
        shutil.copy2(source / "webapp/node_modules/browser-whisper" / name, licenses / name)
    (bundle / "plugin.json").write_text(json.dumps(manifest, indent=2) + "\n")
    with tarfile.open(opts.output, "w:gz", format=tarfile.USTAR_FORMAT) as archive:
        archive.add(bundle, arcname=plugin_id)
