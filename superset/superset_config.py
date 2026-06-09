import os

SECRET_KEY = os.environ.get("SUPERSET_SECRET_KEY", "thisISaSECRET_1234")

SQLALCHEMY_DATABASE_URI = os.environ.get(
    "SQLALCHEMY_DATABASE_URI",
    "postgresql+psycopg2://superset:superset@superset-db/superset",
)

FEATURE_FLAGS = {
    "EMBEDDED_SUPERSET": True,
}

# Guest token config
GUEST_TOKEN_JWT_SECRET = os.environ.get("GUEST_TOKEN_JWT_SECRET", "guest-token-secret-change-me")
GUEST_TOKEN_JWT_ALGO = "HS256"
GUEST_TOKEN_JWT_EXP_SECONDS = 300
GUEST_ROLE_NAME = "Public"

# CORS is handled by the Go backend proxy; Superset itself doesn't need it
# (flask-cors is not bundled in the apache/superset base image)
ENABLE_CORS = False

# Allow embedding in iframes from our frontend
HTTP_HEADERS = {"X-Frame-Options": "ALLOWALL"}

TALISMAN_ENABLED = False

