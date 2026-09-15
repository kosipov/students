"""Yandex Cloud Function that fetches the university's schedule page for the site.

The university doesn't answer requests from the site's server, so the site asks this function instead:
it runs in Yandex Cloud, requests campus.syktsu.ru from a Russian address and returns the page as is.

It is not an open proxy: it only posts the given form to the single address in CAMPUS_URL,
and only for callers that know RELAY_SECRET.

Environment variables of the function:
  RELAY_SECRET - the same value as CAMPUS_RELAY_SECRET of the site; required.
  CAMPUS_URL   - the schedule page; defaults to the teacher schedule of campus.syktsu.ru.
"""

import base64
import hmac
import os
import urllib.error
import urllib.request

CAMPUS_URL = os.environ.get("CAMPUS_URL", "https://campus.syktsu.ru/schedule/teacher/")
RELAY_SECRET = os.environ.get("RELAY_SECRET", "")
# A regular desktop browser's, so requests don't point to the site that makes them.
USER_AGENT = (
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"
)
CAMPUS_TIMEOUT = 20
MAX_REQUEST_BODY = 64 * 1024
MAX_PAGE = 4 * 1024 * 1024


def handler(event, context):
    if event.get("httpMethod") != "POST":
        return _text(405, "Only POST is allowed")

    headers = {key.lower(): value for key, value in (event.get("headers") or {}).items()}
    if not RELAY_SECRET or not hmac.compare_digest(headers.get("x-relay-secret", ""), RELAY_SECRET):
        return _text(403, "Forbidden")

    body = event.get("body") or ""
    if event.get("isBase64Encoded"):
        body = base64.b64decode(body).decode("utf-8", "replace")
    if len(body) > MAX_REQUEST_BODY:
        return _text(413, "Request is too large")

    request = urllib.request.Request(
        CAMPUS_URL,
        data=body.encode("utf-8"),
        headers={
            "Content-Type": "application/x-www-form-urlencoded",
            "User-Agent": USER_AGENT,
            "Accept-Language": "ru-RU,ru;q=0.9",
        },
    )

    try:
        with urllib.request.urlopen(request, timeout=CAMPUS_TIMEOUT) as response:
            page = response.read(MAX_PAGE)
            status = response.status
    except urllib.error.HTTPError as error:
        return _text(502, f"Campus responded {error.code}")
    except Exception as error:  # network errors and timeouts
        return _text(504, f"Campus did not answer: {error}")

    return {
        "statusCode": status,
        "headers": {"Content-Type": "text/html; charset=utf-8"},
        "body": page.decode("utf-8", "replace"),
    }


def _text(status, message):
    return {"statusCode": status, "headers": {"Content-Type": "text/plain; charset=utf-8"}, "body": message}
