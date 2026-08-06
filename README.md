# 🐱 Seppy CLI — Dynamic TUI Next.js Generator Framework

```text
    ▄████▄ ██████ █████▄ █████▄ ██  ██        /\_/\  zZZ
    ██▄▄▄▄ ██▄▄   ██▄▄██ ██▄▄██ ▀████▀       ( -.- )_
    ▄▄▄▄██ ██████ ██     ██       ██         (   "   )_
```

**Seppy** is a lightweight, high-performance Terminal User Interface (TUI) generator framework for scaffolding Next.js applications.

Seppy is **100% dynamic and unopinionated**: it does not force opinionated templates or hardcoded files on users. Instead, `seppy.exe` acts as a clean, dynamic TUI engine that reads skills, markdown docs, and configuration files directly from **each user's personal `~/.seppy` runtime directory**.

---

## ✨ Dynamic Architecture

- 📂 **Personal User Library (`~/.seppy`)**:
  - `~/.seppy/cache/skills/`: Put your own custom agent skills folders here. Seppy will dynamically scan and list them in Step 3.
  - `~/.seppy/docs/`: Put your own custom `.md` architecture files here. Seppy will dynamically scan and list them in Step 4.
  - `~/.seppy/config.json`: Define your own default NPM packages and custom `npx` skill installation commands.
- 🐱 **Interactive Cat Mascot**: Animated ASCII cat companion that types contextual comments, blinks, and enters idle sleep mode (`zZZ`) after 10 seconds of inactivity.
- 🎨 **Responsive Viewport Engine**: Scaling dynamic layout engine (`normal`, `compact`, `veryCompact`) that automatically adapts to small or minimized terminal windows without breaking ANSI cursor positioning.
- ⚡ **Non-Blocking Sound Engine**: Built-in Win32 audio synthesizer for satisfying, subtle UI sound feedback.
- 🖱️ **Full Mouse & Keyboard Support**: Scroll lists smoothly using the mouse wheel, arrow keys, or `Page Up / Down`.

---

## 📋 Prerequisites

Before setting up Seppy, ensure you have:

1. **Windows 10/11** (PowerShell or Windows Terminal).
2. **Node.js & npm / npx** (for running `create-next-app` and installing packages).
3. **Go 1.21+** (REQUIRED to compile the dynamic TUI engine).

---

## 🚀 Quick Setup & Installation

### 1. Run the PowerShell Installer

```powershell
cd path\to\setup
.\install.ps1
```

The installer will automatically:

- Build the Go source code into the `seppy.exe` binary.
- Create `~/.seppy/bin`, `~/.seppy/docs`, and `~/.seppy/cache/skills`.
- Copy compiled `seppy.exe` into `~/.seppy/bin`.
- Clean and update your **User PATH** to prioritize `~/.seppy/bin`.

### 2. Launch Seppy

Reopen your terminal and run:

```powershell
seppy
```

---

## 🎮 Navigation & Keyboard Shortcuts

| Shortcut                 | Action                                                                            |
| :----------------------- | :-------------------------------------------------------------------------------- |
| `[Enter]`                | Confirm selection & proceed to next step                                          |
| `[ESC]`                  | Return to previous setup step                                                     |
| `[Tab]`                  | Toggle inline **Custom Sources** tab (add custom skill commands, markdown docs, NPM packages) |
| `[Ctrl+L]`               | Open **System Locations** modal (configuration & directory paths)                 |
| `[Space]`                | Toggle checkbox selection                                                         |
| `[A]`                    | Select All items in current list                                                  |
| `[N]`                    | Clear All items in current list                                                   |
| `↑` / `↓` or Mouse Wheel | Scroll up/down through lists and viewport content                                 |
| `[Ctrl+C]`               | Cancel setup and force exit CLI                                                   |

---

## ⚙️ Configuration (`~/.seppy/config.json`)

Seppy starts completely clean! It automatically creates a user configuration file at `~/.seppy/config.json`. You can edit this file to pre-configure custom NPM packages or skill installation commands, or add them dynamically via the `[Tab]` menu in the CLI:

```json
{
  "default_unselect_all": true,
  "custom_skills_commands": [],
  "custom_npm_packages": []
}
```

---

## 📄 License

MIT License. Free for open-source and commercial use.
