set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

BOILERPLATE_DIR="docs/user/tutorials"

if [ ! -d "$BOILERPLATE_DIR/snips" ]; then
  echo "boilerplate snippets directory does not exist. creating one..."
  mkdir -p docs/user/tutorials/snips
fi

for f in "$BOILERPLATE_DIR"/*.md; do
  bp_file=$(echo "$f" | awk -F'/' '{ print $NF }' | cut -f1 -d'.')
  bp_func_name=$(echo "$bp_file" | tr '-' '_')
  python3 snip.py "$f" \
      -d docs/user/tutorials/snips \
      -p "bpsnip_$bp_func_name" \
      -f "$bp_file.sh" \
      -b "$BOILERPLATE_DIR/snips"
done

find docs/user/tutorials -name '*.md' -exec grep --quiet '^test: yes$' {} \; -exec python3 scripts/snip.py -b "$BOILERPLATE_DIR/snips" {} \;
