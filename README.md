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
3. **Go 1.21+** (only required if building from source).

---

## 🚀 Quick Setup & Installation

### 1. Run the PowerShell Installer

```powershell
cd path\to\setup
.\install.ps1
```

The installer will automatically:
- Create `~/.seppy/bin`, `~/.seppy/docs`, and `~/.seppy/cache/skills`.
- Copy compiled `seppy.exe` into `~/.seppy/bin`.
- Add `~/.seppy/bin` to your **User PATH**.
- Register the global `seppy` command shortcut in your **PowerShell Profile**.

### 2. Launch Seppy

Reopen your terminal and run:
```powershell
seppy
```

---

## 🎮 Navigation & Keyboard Shortcuts

| Shortcut | Action |
| :--- | :--- |
| `[Enter]` | Confirm selection & proceed to next step |
| `[ESC]` | Return to previous setup step |
| `[Tab]` | Toggle inline **Custom Sources** tab (add custom skill commands / markdown paths) |
| `[Ctrl+L]` | Open **System Locations** modal (configuration & directory paths) |
| `[Space]` | Toggle checkbox selection |
| `[A]` | Select All items in current list |
| `[N]` | Clear All items in current list |
| `↑` / `↓` or Mouse Wheel | Scroll up/down through lists and viewport content |
| `[Q]` or `[Ctrl+C]` | Cancel setup and exit CLI |

---

## ⚙️ Configuration (`~/.seppy/config.json`)

Seppy automatically creates a user configuration file at `~/.seppy/config.json`. You can edit this file to pre-configure custom NPM packages or skill installation commands:

```json
{
  "default_unselect_all": true,
  "custom_skills_commands": [
    {
      "name": "npx skills add (Vercel Find-Skills)",
      "command": "npx skills add https://github.com/vercel-labs/skills --skill find-skills"
    }
  ],
  "custom_npm_packages": [
    "Framer Motion (Animations)",
    "Zustand (State Management)"
  ]
}
```

---

## 📄 License

MIT License. Free for open-source and commercial use.
