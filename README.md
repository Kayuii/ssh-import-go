ssh-import
===========

You're logged onto a cloud instance working on a problem with your fellow devs, and you want to invite them to log in and take a look at these crazy log messages. What do?

Oh. You have to ask them to cat their public SSH key, paste it into IRC (wait, no, it's id\_rsa.pub, not id\_rsa silly!) then you copy it and cat it to the end of authorized\_hosts.

That's where ssh-import comes in. With ssh-import, you can add the public SSH keys from a known, trusted online identity to grant SSH access.

Currently supported identities include Github and Launchpad.

Usage
-----

ssh-import uses short prefix to indicate the location of the online identity. For now, these are:

	'gh:' for Github
	'lp:' for Launchpad

Command line help:

	usage: ssh-import [-h] [-o FILE] USERID [USERID ...]

	Authorize SSH public keys from trusted online identities.

	positional arguments:
  	USERID                User IDs to import

	optional arguments:
  	-h, --help            show this help message and exit
  	-o FILE, --output FILE
                        	Write output to file (default ~/.ssh/authorized_keys)

Example
-------

If you wanted me to be able to ssh into your server, as the desired user on that machine you would use:

	$ ssh-import gh:cmars

You can also import multiple users on the same line, even from different key services, like so:

	$ ssh-import gh:cmars lp:kirkland

Used with care, it's a great collaboration tool!

Installing
----------

ssh-import can be installed on golang >= 1.16 with a recent version of go:

	$ go install github.com/kayuii/ssh-import-go/cmd/ssh-import

Extending
---------

You can add support for your own SSH public key providers by creating a script named ssh-import-*prefix*. Make the script executable and place it in the same bin directory as ssh-import.

The script should accept the identity username for the service it connects to, and output lines in the same format as an ~/.ssh/authorized\_keys file.

If you do develop such a handler, I recommend that you connect to the service with SSL/TLS, and require a valid certificate and matching hostname. Use Requests.get(url, verify=True), for example.

Credits
-------

This repo refers to the following project:

http://launchpad.net/ssh-import-id
