#!/usr/bin/env bash
set -e

# Directory where this script sits
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# backend/scripts → backend
ROOT_DIR="$(realpath "$SCRIPT_DIR/..")"

# Load .env if present
ENV_PATH="$ROOT_DIR/../.env"
if [ -f "$ENV_PATH" ]; then
  echo "🔄 Loading $ENV_PATH"
  set -a
  source "$ENV_PATH"
  set +a
else
  echo "⚠️  No .env found at $ENV_PATH"
fi

# Check DATA_ROOT
if [ -z "$DATA_ROOT" ]; then
  echo "❌ DATA_ROOT env variable not set."
  echo "Add to your .env or export it:"
  echo "    export DATA_ROOT=/opt/nodeherder"
  exit 1
fi

echo "📁 DATA_ROOT = $DATA_ROOT"

USER="mqttuser"
PASS=$(openssl rand -base64 16)


# Directory where this script sits
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# backend/scripts → backend
ROOT_DIR="$(realpath "$SCRIPT_DIR/..")"

# --------------------------------------
# ✅ MQTT folders under DATA_ROOT
# --------------------------------------
CONFIG_DIR="$DATA_ROOT/mqtt/config"
DATA_DIR="$DATA_ROOT/mqtt/data"
LOG_DIR="$DATA_ROOT/mqtt/log"

mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"

# Set permissive permissions so mosquitto container can access
chmod -R 755 "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"

echo "✅ Generating Mosquitto password..."

# Get current user's UID and GID
CURRENT_UID=$(id -u)
CURRENT_GID=$(id -g)

# Create password file using Docker with current user permissions
docker run --rm \
  --user "$CURRENT_UID:$CURRENT_GID" \
  -v "$CONFIG_DIR:/mosquitto/config" \
  eclipse-mosquitto:2 \
  sh -c "touch /mosquitto/config/password.txt && chmod 600 /mosquitto/config/password.txt && mosquitto_passwd -b /mosquitto/config/password.txt '$USER' '$PASS'"

# Ensure password file has secure permissions
chmod 600 "$CONFIG_DIR/password.txt"

echo "✅ Password written → $CONFIG_DIR/password.txt"

# Create mosquitto.conf if missing
MOSQ_CONF="$CONFIG_DIR/mosquitto.conf"
if [ ! -f "$MOSQ_CONF" ]; then
cat <<EOF > "$MOSQ_CONF"
allow_anonymous false
password_file /mosquitto/config/password.txt
listener 1883
EOF
    echo "✅ Created → $MOSQ_CONF"
else
    echo "ℹ️  Existing config found → $MOSQ_CONF"
fi


# --------------------------------------
# ✅ backend/.env stays inside repo
# --------------------------------------
ENV_FILE="$ROOT_DIR/.env"
touch "$ENV_FILE"

# Update/append MQTT_USER
grep -q "^MQTT_USER=" "$ENV_FILE" 2>/dev/null \
  && sed -i "s|^MQTT_USER=.*|MQTT_USER=${USER}|" "$ENV_FILE" \
  || echo "MQTT_USER=${USER}" >> "$ENV_FILE"

# Update/append MQTT_PASS
grep -q "^MQTT_PASS=" "$ENV_FILE" 2>/dev/null \
  && sed -i "s|^MQTT_PASS=.*|MQTT_PASS=${PASS}|" "$ENV_FILE" \
  || echo "MQTT_PASS=${PASS}" >> "$ENV_FILE"

# Build URL
MQTT_URL="tcp://${USER}:${PASS}@mqtt:1883"

# Update/append MQTT_URL
grep -q "^MQTT_URL=" "$ENV_FILE" 2>/dev/null \
  && sed -i "s|^MQTT_URL=.*|MQTT_URL=${MQTT_URL}|" "$ENV_FILE" \
  || echo "MQTT_URL=${MQTT_URL}" >> "$ENV_FILE"


echo ""
echo "✅ MQTT credentials generated"
echo "   USER = $USER"
echo "   PASS = $PASS"
echo ""
echo "✅ Files created/updated:"
echo "   $CONFIG_DIR/password.txt"
echo "   $MOSQ_CONF"
echo "   $ENV_FILE"
echo ""
echo "✅ MQTT_URL written:"
echo "   $MQTT_URL"
echo ""
echo "✅ Done"

######################################################################
# update Zigbee2MQTT configuration.yaml
######################################################################

Z2M_DIR="$DATA_ROOT/zigbee2mqtt-data"
mkdir -p "$Z2M_DIR"

Z2M_CONFIG="$Z2M_DIR/configuration.yaml"

echo ""
echo "🔧 Updating Zigbee2MQTT config: $Z2M_CONFIG"

# Convert tcp:// to mqtt://
TMP="${MQTT_URL/tcp:\/\//mqtt://}"
# remove creds → everything before @
Z2M_MQTT_SERVER="${TMP#*@}"      # remove prefix up to @
Z2M_MQTT_SERVER="mqtt://$Z2M_MQTT_SERVER"

echo "🔗 Zigbee2MQTT MQTT server = $Z2M_MQTT_SERVER"

Z2M_SERIAL_PORT="${Z2M_SERIAL_PORT:-/dev/ttyUSB0}"
echo "🔌 Using Zigbee serial port: $Z2M_SERIAL_PORT"

######################################################################
#  create full template
######################################################################

# if [ -f "$Z2M_CONFIG" ]; then
#     TS=$(date +%s)
#     cp "$Z2M_CONFIG" "$Z2M_CONFIG.bak.$TS"
#     echo "📁 Backed up existing config → $Z2M_CONFIG.bak.$TS"
# fi
cat <<EOF > "$Z2M_CONFIG"
permit_join: true

mqtt:
  base_topic: zigbee2mqtt
  server: ${Z2M_MQTT_SERVER}
  user: ${USER}
  password: ${PASS}

serial:
  port: ${Z2M_SERIAL_PORT}

frontend:
  port: 8080

advanced:
  network_key: GENERATE
  last_seen: ISO_8601_local
EOF

echo "✅ Zigbee2MQTT configuration written fresh"

echo "✅ Zigbee2MQTT configuration updated"
echo ""
echo "✅ Done"