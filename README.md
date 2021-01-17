# What the
Spad-mats is swedish and translates into Mats (a common Swedish male name) with the shovel.

Mats is here to dig out dirty git-repository secrets and selfishly expose them to you as the honest hardworking swede he is.

Here's a picture of someone that could be named Mats. Let's pretend he is.

![spad-mats-and-his-dog-loffe](assets/spad-mats.jpg)

# Prerequisites
You need:
* [Go](https://golang.org/) installed and on your path.
* [Git](https://git-scm.com/) installed and on your path (since this program uses exec.Command to run git commands).
* [Postgres](https://www.postgresql.org/) database installed and running.

Tip: this is what [sider](https://github.com/jonaslu/sider) is made for - quick experiments with no hassle of a full database installation.

In addition you need the psql command-line. It's available as a standalone package in most distros so you don't need the full postgres-installation (e g [postgres-libs](https://archlinux.org/packages/extra/x86_64/postgresql-libs/) in arch linux).

If you want to import the 100 most popular repos via [clone-repos.sh](clone-repos.sh) and run the stats in [RESULT.md](RESULT.md) then you also need [pup](https://github.com/ericchiang/pup) installed. Your package manager should have it.

# Installation
Clone this repository to a local folder.

Run ./db-setup.sh with the PG_DSN environment set. You can either set the environment variable PG_DSN to a connection string of your choice, or you can accept the default of postgres://postgres@localhost:5432/spad-mats?sslmode=disable.

Postgres must be running when you do this and the script relies on bash and psql.

# cmd/import/main.go
Imports the commit-log of a local git-repository into two tables in the database.
If the repository is above a certain commit-count, it uses sampling to lower the number of commits.

It uses two environment-variables: PG_DSN for the connection string (or you can use the default) and COUNT which is the number above it starts to sample commits. If not set will sample the entire repo.

Usage:
`go run cmd/import/main.go <path-to-repo> <repo-name>`

It'll consider the currently checked out branch of the given repo (if bare cloned will default to master or main if newer github repository).

It is assumed that the git-repo exists on the correct path and that it contains at least 1 commit. What happens otherwise is undefined.

# clone-repos.sh
Script that pulls the 100 most popular repos from github and then imports them into the
postgres database with the command above. Samples repositories down to a 1000 commits.

It uses your /tmp/ folder to clone repos into and then deletes them. Since some repos are quite
large (such as linux, go and nodejs) it takes one positional parameter which is a folder to clone into
should your /tmp/ partition be too small.

This script requires you to have git and pup installed and on your path.

# RESULT.md
A discussion relating to my blog-post: [link](link) on which I use the ./clone-repos.sh result in postgres to check the atomicity and literacy of the 100 most popular repos on github.
