#!/bin/sh

SERVER="__SERVER_URL__"
TEXT="$1"

DB="/mnt/onboard/.kobo/KoboReader.sqlite"
LOG="/mnt/onboard/.adds/nm/recap-scripts/last-selection.log"
SQLITE="$(command -v sqlite3 2>/dev/null)"
WGET="$(command -v wget 2>/dev/null)"

[ -z "$SQLITE" ] && SQLITE="/mnt/onboard/.adds/nm/programs/sqlite3"
[ -z "$WGET" ] && WGET="/usr/bin/wget"

mkdir -p /mnt/onboard/.adds/nm/recap-scripts
printf '%s\n' "---- $(date) ----" > "$LOG"
echo "Server: $SERVER" >> "$LOG"
echo "Arguments: $#" >> "$LOG"

if [ -z "$TEXT" ]; then echo "No selected text was passed to the script." | tee -a "$LOG"; exit 5; fi
if [ ! -x "$WGET" ]; then echo "wget not found at $WGET" | tee -a "$LOG"; exit 3; fi
echo "wget: $WGET" >> "$LOG"

urlencode_all() {
  printf "%s" "$1" | od -An -tx1 | tr -d " \n" | sed 's/../%&/g'
}

TITLE=""
AUTHOR=""
echo "sqlite path: $SQLITE" >> "$LOG"
if [ -x "$SQLITE" ]; then
  echo "sqlite executable: yes" >> "$LOG"
  "$SQLITE" -version >> "$LOG" 2>&1
else
  echo "sqlite executable: no" >> "$LOG"
fi

if [ -f "$DB" ]; then
  echo "database exists: yes" >> "$LOG"
else
  echo "database exists: no" >> "$LOG"
fi

if [ -r "$DB" ]; then
  echo "database readable: yes" >> "$LOG"
else
  echo "database readable: no" >> "$LOG"
fi

if [ -x "$SQLITE" ] && [ -r "$DB" ]; then
  TAB="$(printf '\t')"
  echo "Recent book candidates:" >> "$LOG"
  "$SQLITE" -cmd ".timeout 3000" -separator " | " "$DB" \
    "SELECT ContentType,
            COALESCE(Title, ''),
            COALESCE(Attribution, ''),
            COALESCE(DateLastRead, '')
       FROM content
      WHERE DateLastRead IS NOT NULL
        AND COALESCE(Title, '') <> ''
      ORDER BY DateLastRead DESC
      LIMIT 5;" >> "$LOG" 2>&1
  CANDIDATE_RC=$?
  echo "candidate query exit code: $CANDIDATE_RC" >> "$LOG"

  SQL="SELECT replace(replace(replace(COALESCE(Title,''), char(10), ' '), char(13), ' '), char(9), ' '),
              replace(replace(replace(COALESCE(Attribution,''), char(10), ' '), char(13), ' '), char(9), ' ')
       FROM content
       WHERE ContentType=6 AND DateLastRead IS NOT NULL
       ORDER BY DateLastRead DESC
       LIMIT 1;"
  ROW="$($SQLITE -cmd ".timeout 3000" -separator "$TAB" "$DB" "$SQL" 2>>"$LOG")"
  QUERY_RC=$?
  echo "title query exit code: $QUERY_RC" >> "$LOG"
  echo "title query returned chars: ${#ROW}" >> "$LOG"
  IFS="$TAB" read -r TITLE AUTHOR <<EOF2
$ROW
EOF2
fi

[ -n "$TITLE" ] && echo "Database title: $TITLE" >> "$LOG"
[ -z "$TITLE" ] && TITLE="Kobo selection"

BODY="title=$(urlencode_all "$TITLE")"
# BODY="$BODY&author=$(urlencode_all "$AUTHOR")"
BODY="$BODY&text=$(urlencode_all "$TEXT")"

echo "Title: $TITLE" >> "$LOG"
echo "Selection chars: ${#TEXT}" >> "$LOG"

echo "Uploading..." >> "$LOG"
$WGET -O - --post-data="$BODY" "$SERVER/api/selection" >> "$LOG" 2>&1
RC=$?
echo "wget exit code: $RC" >> "$LOG"
if [ "$RC" -ne 0 ]; then echo "Upload failed with exit code $RC" | tee -a "$LOG"; exit "$RC"; fi

echo "Uploaded selected text."
exit 0
