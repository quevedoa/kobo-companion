#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
MOUNT_PATH=${KOBO_MOUNT:-/media/qvdo/KOBOeReader}
server_url=${SERVER_URL:-https://qvdo.taila7ecf5.ts.net}
server_url=${server_url%/}
DEST_ROOT="$MOUNT_PATH/.adds/nm"
DEST_SCRIPTS="$DEST_ROOT/recap-scripts"
DEST_PROGRAMS="$DEST_ROOT/programs"

if [ ! -d "$MOUNT_PATH" ]; then
  echo "Kobo is not mounted at $MOUNT_PATH." >&2
  exit 1
fi

if [ ! -d "$MOUNT_PATH/.adds" ]; then
  echo "Missing $MOUNT_PATH/.adds; is this the Kobo volume?" >&2
  exit 1
fi

echo "Kobo:   $MOUNT_PATH"
echo "Server: $server_url"
printf "Update NickelMenu config, programs, and recap scripts? [y/N] "
read -r answer
case "$answer" in
  y|Y|yes|YES) ;;
  *) echo "Cancelled."; exit 0 ;;
esac

staging_dir=$(mktemp -d "${TMPDIR:-/tmp}/kobo-companion-sync.XXXXXX")
trap 'rm -rf "$staging_dir"' EXIT HUP INT TERM

sed "s|__SERVER_URL__|$server_url|g" \
  "$SCRIPT_DIR/nickelmenu/nickelmenu-config-template" \
  > "$staging_dir/config"

mkdir -p "$staging_dir/scripts"
script_count=0
for source_script in "$SCRIPT_DIR"/scripts/*; do
  [ -f "$source_script" ] || continue
  script_name=$(basename "$source_script")
  sed "s|__SERVER_URL__|$server_url|g" \
    "$source_script" \
    > "$staging_dir/scripts/$script_name"
  sh -n "$staging_dir/scripts/$script_name"
  script_count=$((script_count + 1))
done

if [ "$script_count" -eq 0 ]; then
  echo "No Kobo scripts found in $SCRIPT_DIR/scripts." >&2
  exit 1
fi

mkdir -p "$DEST_SCRIPTS" "$DEST_PROGRAMS"

cp "$staging_dir/config" "$DEST_ROOT/config"
chmod 700 "$DEST_ROOT/config"

cmp -s "$staging_dir/config" "$DEST_ROOT/config"
for staged_script in "$staging_dir"/scripts/*; do
  script_name=$(basename "$staged_script")
  destination="$DEST_SCRIPTS/$script_name"
  cp "$staged_script" "$destination"
  chmod 700 "$destination"
  cmp -s "$staged_script" "$destination"
done

for source_program in "$SCRIPT_DIR"/programs/*; do
  [ -f "$source_program" ] || continue
  program_name=$(basename "$source_program")
  destination="$DEST_PROGRAMS/$program_name"
  cp "$source_program" "$destination"
  chmod 700 "$destination"
  cmp -s "$source_program" "$destination"
done

echo "Kobo files updated successfully."
echo "Safely eject and reboot the Kobo to reload NickelMenu."
