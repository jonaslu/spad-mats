# What the
Spad-mats is swedish and translates into Mats (a common Swedish male name) with the shovel.

Mats is here to dig out dirty git-repository secrets and selfishly expose them to you as the honest hardworking swede he is.

Here's a picture of someone that could be named Mats. Let's pretend he is.

![spad-mats-and-his-dog-loffe](assets/spad-mats.jpg)

# Prerequisites
You need a [postgres](https://www.postgresql.org/) database installed and running.

Tip: this is what [sider](https://github.com/jonaslu/sider) is made for - quick experiments with no hassle of a full database installation.

In addition you need the psql command-line. It's available as a standalone package in most distros so you don't need the full postgres-installation (e g [postgres-libs](https://archlinux.org/packages/extra/x86_64/postgresql-libs/) in arch linux).

# Installation
Clone this repository to a local folder.

Run ./db-setup.sh with the PG_DSN environment set. You can either set the environment variable PG_DSN to a connection string of your choice, or you can accept the default of postgres://postgres@localhost:5432/spad-mats?sslmode=disable.

Postgres must be running when you do this and the script relies on bash and psql.
