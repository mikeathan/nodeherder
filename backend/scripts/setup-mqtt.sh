#!/usr/bin/env bash
set -e

USER="mqttuser"
PASS=$(openssl rand -base64 16)

# Directory where this script sits
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# backend/scripts → backend
ROOT_DIR="$(realpath "$SCRIPT_DIR/..")"

# backend/mqtt dirs
CONFIG_DIR="$ROOT_DIR/mqtt/config"
DATA_DIR="$ROOT_DIR/mqtt/data"
LOG_DIR="$ROOT_DIR/mqtt/log"

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

# backend/.env
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


# --------------------------------------
# ✅ Update Zigbee2MQTT configuration.yaml
# --------------------------

# TODO:
# That can be parameterized 
Z2M_DIR="$ROOT_DIR/../../zigbee2mqtt-data"
if [ ! -d "$Z2M_DIR" ]; then
    echo "📁 Creating Zigbee2MQTT data directory → $Z2M_DIR"
    mkdir -p "$Z2M_DIR"
fi

Z2M_CONFIG="$Z2M_DIR/configuration.yaml"

echo ""
echo "🔧 Updating Zigbee2MQTT config at: $Z2M_CONFIG"

# Create file if missing
if [ ! -f "$Z2M_CONFIG" ]; then
    echo "⚠️  No configuration.yaml found — creating a minimal one"
    cat <<EOF > "$Z2M_CONFIG"
mqtt:
  server: mqtt://mqtt:1883
  user: ${USER}
  password: ${PASS}
EOF
else
    echo "✅ Updating existing Z2M configuration.yaml"

    # Ensure mqtt block exists
    if ! grep -q "^mqtt:" "$Z2M_CONFIG"; then
        echo "" >> "$Z2M_CONFIG"
        echo "mqtt:" >> "$Z2M_CONFIG"
    fi

    # server
    if grep -q "server:" "$Z2M_CONFIG"; then
        sed -i "s|server:.*|server: mqtt://mqtt:1883|" "$Z2M_CONFIG"
    else
        sed -i "/^mqtt:/a\  server: mqtt://mqtt:1883" "$Z2M_CONFIG"
    fi

    # user
    if grep -q "user:" "$Z2M_CONFIG"; then
        sed -i "s|user:.*|user: ${USER}|" "$Z2M_CONFIG"
    else
        sed -i "/^mqtt:/a\  user: ${USER}" "$Z2M_CONFIG"
    fi

    # password
    if grep -q "password:" "$Z2M_CONFIG"; then
        sed -i "s|password:.*|password: ${PASS}|" "$Z2M_CONFIG"
    else
        sed -i "/^mqtt:/a\  password: ${PASS}" "$Z2M_CONFIG"
    fi
fi

echo "✅ Zigbee2MQTT MQTT credentials updated"

