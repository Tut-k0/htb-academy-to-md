# HackTheBox Academy to Markdown
This is a simple CLI application that will fetch and convert a HackTheBox Academy module into a local file in Markdown format.
This program will only grab one module at a time, and requires authenticating with the platform. 
You will also need to have the module unlocked, which should go without saying.

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
These steps have changed slightly with the reCaptcha update on the HackTheBox platform. 
I see this current state as a workaround for not dealing with the reCaptcha until I have more time to dig into that.

Essentially instead of passing your email and password, you will just pass your authenticated session cookies to the application to use. 
So the one added step for the workaround is manually logging into the academy (I would assume you are logged into Academy anyway to get the module URL), and extracting your cookies from your browser.
You can fetch these with the developer tools, burp, or a browser extension, whatever works easiest for you.
The cookies will get passed to the new `-c` argument, and you no longer need to pass an email or password.
```bash
# Get the help menu displayed
htb-academy-to-md -h

# Feed the URL to the module.
htb-academy-to-md -m https://academy.hackthebox.com/module/112/section/1060 -c "htb_academy_session=value; XSRF-TOKEN=value; some-other-cookie=value"

# Save images in module locally.
htb-academy-to-md -m https://academy.hackthebox.com/module/112/section/1060 -local_images -c "htb_academy_session=value; XSRF-TOKEN=value; some-other-cookie=value"

# You can also grab multiple modules using a simple loop if preferred. (bash example)
for i in $(cat modules.txt);do htb-academy-to-md -m $i -c "htb_academy_session=value; XSRF-TOKEN=value; some-other-cookie=value";done
```

### Building
```bash
# Run from inside the /src folder.
go build -o htb-academy-to-md
```