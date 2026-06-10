# pitara

Backup and restore your developer environment — language runtimes and global CLI tools.

```bash
npm install -g pitara

pitara scan -o snapshot.json          # capture this machine
pitara restore --from snapshot.json   # rebuild it elsewhere
```

This package is a thin installer: on install it downloads the prebuilt
[Pitara](https://github.com/sailingsam/pitara) binary for your OS and CPU from
the matching GitHub Release. The CLI itself is written in Go.

Supported platforms: macOS, Linux, Windows (x64 and arm64).

Full documentation and source: **https://github.com/sailingsam/pitara**
