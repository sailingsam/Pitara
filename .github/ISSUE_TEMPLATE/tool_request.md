---
name: Tool / plugin request
about: Ask for a runtime or package manager Pitara should support
title: "[tool] Support "
labels: enhancement, good first issue
---

**Which tool?**
Name of the runtime or package manager (e.g. Python, Rust, Ruby, Deno, Homebrew).

**How do you check its version?**
```bash
# e.g. python --version  →  Python 3.12.3
```

**How do you list its global packages / installed version?** (if applicable)
```bash
# e.g. pip list --format=json   /   cargo install --list
```
Paste the **raw output** if you can — it's what a plugin parses.

**Want to add it yourself?**
Adding a tool is one small plugin — see [CONTRIBUTING.md](../blob/main/CONTRIBUTING.md). Happy to guide you through your first PR.
