# HackTheBox Academy to Markdown
This is a simple CLI application that will fetch and convert a HackTheBox Academy module into a local file in Markdown format.
This program only will grab one module at a time, and requires authenticating with the platform. You will also need to have the module unlocked, which should go without saying.

### Running the Application
```bash
# Feed the URL to the first page of the module.
htb-academy-to-md -m https://academy.hackthebox.com/module/112/ -e <email> -p <password>
```
