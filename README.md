# Weaviate Contextionary <img alt='Weaviate logo' src='https://raw.githubusercontent.com/semi-technologies/weaviate/19de0956c69b66c5552447e84d016f4fe29d12c9/docs/assets/weaviate-logo.png' width='180' align='right' />

> The contextionary powers the semantic, context-based searches in weaviate.

Not intended for stand-alone use, use through [weaviate - the decentralized
knowledge graph](https://github.com/semi-technologies/weaviate).

## Versioning

The version tag is `<language-of-db><semver-of-db>-v<semver-of-app>`. So for
example the app version `0.1.0` deployed with the [contextionary vector db
version](https://c11y.semi.technology/contextionary.json) `0.6.0` of the
English language  will have the version `en0.6.0-v0.1.0`. This also
corresponds to the docker tag.

## Languages

Currently available languages include:
* `en` 
* `de`
* `nl`
* `cs`
* `it`

Other languages coming soon.

## Docker Requirements

The build pipeline makes use of Docker's `buildx` for multi-arch builds. Make
sure you run a Docker version which supports `buildx` and have run `docker
buildx create --use` at least once.

## How to build and test project

1. Regenerate schema:

```bash
./gen_proto_code.sh
```

2. Build image:

```bash
LANGUAGE=en MODEL_VERSION=0.16.0 ./build.sh
```

3. Run journey tests:

```bash
LANGUAGE=en MODEL_VERSION=0.16.0 ./build.sh && DIMENSIONS=300 ./test/journey.sh
```
