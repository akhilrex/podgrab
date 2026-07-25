<p align="center">
  <!-- <a href="https://github.com/akhilrex/podgrab">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a> -->

  <h1 align="center" style="margin-bottom:0px">Podgrab-S3</h1>
  <p align="center">Current Version - 2026.07.25</p>

  <p align="center">
    A self-hosted podcast manager to download episodes as soon as they become live
    <br />
    <a href="https://github.com/akhilrex/podgrab"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <!-- <a href="https://github.com/akhilrex/podgrab">View Demo</a>
    · -->
    <a href="https://github.com/akhilrex/podgrab/issues">Report Bug</a>
    ·
    <a href="https://github.com/akhilrex/podgrab/issues">Request Feature</a>
        ·
    <a href="Screenshots.md">Screenshots</a>
  </p>
</p>

> **Podgrab-S3** is a fork of [akhilrex/podgrab](https://github.com/akhilrex/podgrab) that adds an S3-compatible object storage backend (AWS S3, MinIO, DigitalOcean Spaces, etc.) as an alternative to the local `/assets` volume for storing downloaded episodes and images. All credit for the original application goes to [Akhil Gupta](https://github.com/akhilrex) - see [Contact](#contact) and [License](#license).

<!-- TABLE OF CONTENTS -->

## Table of Contents

- [About the Project](#about-the-project)
  - [Motivation](#motivation)
  - [Built With](#built-with)
  - [Features](#features)
- [Installation](#installation)
  - [Migrating from Local Storage to S3](#migrating-from-local-storage-to-s3)
- [License](#license)
- [Roadmap](#roadmap)
- [Contact](#contact)

<!-- ABOUT THE PROJECT -->

## About The Project

Podgrab is a is a self-hosted podcast manager which automatically downloads latest podcast episodes. It is a light-weight application built using GO.

It works best if you already know which podcasts you want to monitor. However there is a podcast search system powered by iTunes built into Podgrab

### Motivation

Podgrab started as a tool that initially built to solve a specific problem akhilrex had. During the COVID pandemic times he started going for a run. He did not prefer taking his phone along so he would add podcast episodes to his smart watch which could be connected with his bluetooth earphones. Most podcasting apps do not expose the mp3 files directly which is why he decided to build this quick tool for himself. Once it reached a stage where his requirements were fulfilled he decided to make it a little pretty and share it with everyone else.

![Product Name Screen Shot][product-screenshot]
[More Screenshots](Screenshots.md)

### Built With

- [Go](https://golang.org/)
- [Go-Gin](https://github.com/gin-gonic/gin)
- [GORM](https://github.com/go-gorm/gorm)
- [SQLite](https://www.sqlite.org/index.html)

### Features
- S3-compatible object storage backend for episodes/images (AWS S3, MinIO, DigitalOcean Spaces, etc.), as an alternative to local storage
- Download/Archive complete podcast
- Auto-download new episodes
- Tag/Label podcasts into groups
- Download on demand
- Podcast Discovery - Search and Add podcasts using iTunes API
- Full-fledged podcast player - Play downloaded files or stream from original source. Play single episodes, full podcasts and podcast groups(tags)
- Add using direct RSS feed URL / OMPL import / Search
- Basic Authentication
- Existing episode file detection - Prevent re-downloading files if already present
- Easy OPML import/export
- Customizable episode names
- Dark Mode
- Self Hosted / Open Source
- Docker support

## Installation

The easiest way to run Podgrab is to run it as a docker container.

### Using Docker

Simple setup without mounted volumes (for testing and evaluation)

```sh
  docker run -d -p 8080:8080 --name=podgrab vanityshed/podgrab-s3
```

Binding local volumes to the container

```sh
   docker run -d -p 8080:8080 --name=podgrab -v "/host/path/to/assets:/assets" -v "/host/path/to/config:/config"  vanityshed/podgrab-s3
```

### Using Docker-Compose

Modify the docker compose file provided [here](https://github.com/stevesbrain/podgrab-s3/blob/master/docker-compose.yml) to update the volume and port binding and run the following command

```yaml
version: "2.1"
services:
  podgrab:
    image: vanityshed/podgrab-s3:latest
    container_name: podgrab
    environment:
      - CHECK_FREQUENCY=240
     # - PASSWORD=password     ## Uncomment to enable basic authentication, username = podgrab
      # Uncomment to store episodes/images in S3 instead of the /assets volume,
      # and have the RSS feed link directly to the bucket instead of proxying.
      # - STORAGE_BACKEND=s3
      # - S3_BUCKET=my-podgrab-bucket
      # - S3_REGION=us-east-1
      # - S3_ACCESS_KEY_ID=xxxxxxxx
      # - S3_SECRET_ACCESS_KEY=xxxxxxxx
      # - S3_ACL=public-read       ## only needed for buckets/providers that still use ACLs
      # - S3_PUBLIC_URL_BASE=https://cdn.example.com   ## optional, e.g. a CloudFront domain
    volumes:
      - /path/to/config:/config
      - /path/to/data:/assets
    ports:
      - 8080:8080
    restart: unless-stopped
```

```sh
   docker-compose up -d
```
### Build from Source / Ubuntu Installation

Although personally he feels that using the docker container is the best way of using and enjoying something like Podgrab, a lot of people in the community are still not comfortable with using Docker and wanted to host it natively on their Linux servers. Follow the link below to get a guide on how to build Podgrab from source.

[Build from source / Ubuntu Guide](docs/ubuntu-install.md)
### Environment Variables

| Name            | Description                                                             | Default |
| --------------- | ----------------------------------------------------------------------- | ------- |
| CHECK_FREQUENCY | How frequently to check for new episodes and missing files (in minutes) | 30      |
| PASSWORD        | Set to some non empty value to enable Basic Authentication, username `podgrab`|(empty)|
| PORT            | Change the internal port of the application. If you change this you might have to change your docker configuration as well | (empty) |  
| STORAGE_BACKEND | Where downloaded episodes/images are stored: `local` (the `/assets` volume) or `s3` | local |
| S3_BUCKET       | Bucket name to store episodes/images in. Required when `STORAGE_BACKEND=s3` | (empty) |
| S3_REGION       | AWS region of the bucket | us-east-1 |
| S3_ACCESS_KEY_ID | Access key for the bucket. If unset, falls back to the default AWS credential chain (IAM role, shared credentials file, etc.) | (empty) |
| S3_SECRET_ACCESS_KEY | Secret key paired with `S3_ACCESS_KEY_ID` | (empty) |
| S3_ENDPOINT     | Custom endpoint for S3-compatible providers (MinIO, DigitalOcean Spaces, etc), as a full URL e.g. `https://nyc3.digitaloceanspaces.com` or `http://minio:9000`. A bare hostname is accepted too and assumed to be `https://`. Leave unset for AWS S3 | (empty) |
| S3_USE_PATH_STYLE | Set to `true` if your provider requires path-style requests (`https://endpoint/bucket/key`) instead of virtual-hosted style | false |
| S3_KEY_PREFIX   | Optional prefix (subfolder) under which all objects are stored in the bucket | (empty) |
| S3_ACL          | Object ACL applied on upload, e.g. `public-read`. Leave unset for buckets with ACLs disabled (the current AWS default) and control public access with a bucket policy instead | (empty) |
| S3_PUBLIC_URL_BASE | Override the base URL used to build RSS/enclosure links, e.g. a CloudFront/CDN domain fronting the bucket. If unset, the standard `https://<bucket>.s3.<region>.amazonaws.com` (or endpoint-derived) URL is used | (empty) |

When `STORAGE_BACKEND=s3`, downloaded episode and image files are uploaded straight to the bucket and the RSS feed links directly to the bucket (or `S3_PUBLIC_URL_BASE`) URL instead of proxying the bytes through Podgrab — the bucket (or fronting CDN) must therefore make those objects publicly readable. The `/assets` volume is only needed for `local` storage.

### Migrating from Local Storage to S3

If you already have a Podgrab install using the local `/assets` volume, moving to S3 storage takes two parts: copying the files, then updating the database records that point at them. `download_path`/`local_image` in `podcast_items` are stored as plain relative paths under `/assets` for local storage, and as bucket-relative object keys for S3 storage — so the migration is just stripping that `/assets/` prefix once the files themselves live in the bucket at the same relative layout.

1. **Stop the container** before making any of these changes, to avoid the app writing new files mid-migration.
2. **Copy your existing data to the bucket, preserving the folder structure**, e.g. with [rclone](https://rclone.org/):
   ```sh
   rclone sync /path/to/assets remote:my-podgrab-bucket
   ```
   The result should mirror what's under `/assets` today — one folder per podcast, with episode/image files inside — directly at the bucket root (no extra prefix folder, unless you intend to set `S3_KEY_PREFIX` and account for it below).
3. **Back up `podgrab.db`** before editing it (`cp podgrab.db podgrab.db.bak`).
4. **Run the migration scripts** in [`scripts/`](scripts/) against `podgrab.db` with the [sqlite3](https://www.sqlite.org/cli.html) CLI:
   ```sh
   sqlite3 podgrab.db < scripts/preview-local-paths-to-s3-keys.sql   # read-only, sanity-check first
   sqlite3 podgrab.db < scripts/migrate-local-paths-to-s3-keys.sql   # applies the update
   ```
5. **Update your compose file** to set `STORAGE_BACKEND=s3` and the related `S3_*` variables from the table above. Leave `S3_KEY_PREFIX` unset unless you deliberately synced your files under a prefix in step 2 — the scripts store bare relative keys, and any configured prefix is added back on at read time, so setting one without a matching prefix in the bucket will break every link.
6. **Start the container again** and confirm episodes play and download correctly before removing the local `/assets` copy.

### Setup

- Enable *websocket support* if running behind a reverse proxy. This is needed for the "Add to playlist" functionality.
- Go through the settings page once and change relevant settings before adding podcasts.

## License

Distributed under the GPL-3.0 License. See `LICENSE` for more information.

## Roadmap

- [x] Basic Authentication
- [x] Append Date to filename
- [x] iTunes Search
- [x] Existing episodes detection (Will not redownload if files exist even with a fresh install)
- [x] Downloading/downloaded indicator
- [x] Played/Unplayed Flag
- [x] OPML import
- [x] OPML export
- [x] In built podcast player
- [ ] Set ID3 tags if not set 
- [X] Filtering and Sorting options
- [ ] Native installer for Windows/Linux/MacOS




<!-- CONTACT -->

## Contact

Akhil Gupta - [@akhilrex](https://twitter.com/akhilrex)

Project Link: [https://github.com/stevesbrain/podgrab-3](https://github.com/akhilrex/podgrab)

<a href="https://www.buymeacoffee.com/akhilrex" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->

[contributors-shield]: https://img.shields.io/github/contributors/akhilrex/podgrab.svg?style=flat-square
[contributors-url]: https://github.com/akhilrex/podgrab/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/akhilrex/podgrab.svg?style=flat-square
[forks-url]: https://github.com/akhilrex/podgrab/network/members
[stars-shield]: https://img.shields.io/github/stars/akhilrex/podgrab.svg?style=flat-square
[stars-url]: https://github.com/akhilrex/podgrab/stargazers
[issues-shield]: https://img.shields.io/github/issues/akhilrex/podgrab.svg?style=flat-square
[issues-url]: https://github.com/akhilrex/podgrab/issues
[license-shield]: https://img.shields.io/github/license/akhilrex/podgrab.svg?style=flat-square
[license-url]: https://github.com/akhilrex/podgrab/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=flat-square&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/akhilrex
[product-screenshot]: images/screenshot.jpg
