# HackTheBox Academy to Markdown
This is a simple CLI application that will fetch and convert a HackTheBox Academy module into a local file in Markdown format.
This program will only grab one module at a time, and requires authenticating with the platform. 
You will also need to have the module unlocked, which should go without saying.

**Updated for HTB Academy 2.0** - This tool has been updated to work with the latest HTB Academy platform using the new API endpoints.

I personally use [Obsidian](https://obsidian.md/) as my note-taking tool, and this application is tailored and tested for rendering markdown utilizing it.
Most other note-taking tools that can import markdown files should work fine as well.

### Disclaimer
**Please note that this application is not intended for use in uploading or sharing the end result content.**
The application is solely designed for personal use and any content created using this application should not be shared or uploaded to any platform without proper authorization and consent from HackTheBox.
The contributors of this application are not responsible for any unauthorized use or distribution of the content created using this application.

### Installing
Check the releases folder [here](https://github.com/Tut-k0/htb-academy-to-md/releases), and download the most recent executable for your operating system.
All the executables listed here are for x64 and amd64. If there is not an executable for your OS or architecture, you can simply build the application. (See building section below.)

### Running
This application uses session cookie authentication rather than email/password login. This approach is necessary because:
- HTB Academy can implement reCAPTCHA on login forms
- Two-factor authentication (2FA) may be required for accounts

Since automating reCAPTCHA and 2FA programmatically is not feasible, the session cookie method provides a simple workaround.
Simply log into HTB Academy manually (as you normally would to access modules), extract your session cookie from your browser, and pass it to the application.

You can extract cookies using browser developer tools (F12 → Application/Storage → Cookies), Burp Suite, or a browser extension.
The session cookie you need is `htb_academy_session`.

```bash
# Get the help menu displayed
htb-academy-to-md -h

# Feed the URL to the module.
htb-academy-to-md -m https://academy.hackthebox.com/module/112/section/1060 -c "htb_academy_session=value"

# Save images in module locally.
htb-academy-to-md -m https://academy.hackthebox.com/module/112/section/1060 -local_images -c "htb_academy_session=value"

# You can also grab multiple modules using a simple loop if preferred. (bash example)
for i in $(cat modules.txt);do htb-academy-to-md -m $i -c "htb_academy_session=value";done
```

### Building
```bash
# Run from inside the /src folder.
go build -o htb-academy-to-md
```