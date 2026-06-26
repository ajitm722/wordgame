#!/usr/bin/env bash

# Clean up any previous server lingering on port 1337
kill $(lsof -ti :1337) 2>/dev/null || true

echo "\$ make test-cover-html"
make test-cover-html 2>/dev/null || true
sleep 2

echo ""
echo "\$ make smoke"
make smoke
sleep 2

echo ""
echo "\$ make run &"
make run &
SERVER_PID=$!
sleep 2

echo ""
echo "\$ make new-game"
NEW_GAME=$(make new-game)
echo "$NEW_GAME"
GAME_ID=$(echo "$NEW_GAME" | jq -r '.id')
sleep 2

echo ""
echo "# Guessing A-Z until game ends"
sleep 1
for LETTER in {A..Z}; do
    echo ""
    echo "\$ make guess ID=$GAME_ID GUESS=$LETTER"
    RESPONSE=$(make guess ID="$GAME_ID" GUESS="$LETTER" 2>&1) || true
    echo "$RESPONSE"
    if echo "$RESPONSE" | jq -e '.word' >/dev/null 2>&1; then
        break
    fi
    sleep 1
done
sleep 2

echo ""
echo "\$ make guess ID=$GAME_ID GUESS=Z"
make guess ID="$GAME_ID" GUESS="Z" 2>&1 || true
sleep 2

echo ""
echo "\$ kill $SERVER_PID"
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true
