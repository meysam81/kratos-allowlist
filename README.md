# kratos-allowlist

Domain validation webhook for Ory Kratos registration flows.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Usage](#usage)
  - [Kratos Configuration](#kratos-configuration)
  - [email.jsonnet](#emailjsonnet)
  - [Run](#run)
- [Environment Variables](#environment-variables)
- [Endpoints](#endpoints)
- [Request](#request)
- [Response](#response)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Usage

### Kratos Configuration

```yaml
selfservice:
  flows:
    registration:
      after:
        password:
          hooks:
            - config:
                body: https://gist.github.com/meysam81/9b6b63d0530987a9236d43d21cbec713/raw/513535c2c30eade74537d85d757edd4c1fc18b73/email.jsonnet
                method: POST
                response:
                  parse: true
                url: http://localhost:8080/v1/validate
              hook: web_hook
```

### email.jsonnet

```jsonnet
function(ctx) {
  email: ctx.identity.traits.email,
}
```

### Run

```bash
export ALLOWED__DOMAINS="example.com company.org"
docker run --rm --name allowlist -e ALLOWED__DOMAINS -dp 8080:8080 ghcr.io/meysam81/kratos-allowlist
```

## Environment Variables

- `ALLOWED__DOMAINS`: Space-separated list of allowed email domains
- `PORT`: Server port (default: 8080)

## Endpoints

- `POST /v1/validate` - Webhook validation endpoint

## Request

```json
{
  "email": "john.doe@example.com"
}
```

## Response

Returns `200 OK` for allowed domains, `400 Bad Request` with validation error for blocked domains.
