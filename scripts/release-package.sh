#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_dir="${repo_root}/dist"
goos="$(go env GOOS)"
goarch="$(go env GOARCH)"
artifact_name="daymine-${goos}-${goarch}.tgz"
binary_dir="${dist_dir}/daymine-${goos}-${goarch}"

rm -rf "${dist_dir}"
mkdir -p "${binary_dir}"

npm --prefix "${repo_root}/apps/web" ci --no-audit
npm --prefix "${repo_root}/apps/web" run build

go -C "${repo_root}" build -o "${binary_dir}/daymine" ./apps/daymine/cmd/daymine

cat > "${dist_dir}/release-manifest.json" <<EOF
{
  "repository": "${GITHUB_REPOSITORY:-local}",
  "git_sha": "${GITHUB_SHA:-$(git -C "${repo_root}" rev-parse HEAD 2>/dev/null || echo unknown)}",
  "generated_at_utc": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "artifact": "${artifact_name}",
  "goos": "${goos}",
  "goarch": "${goarch}",
  "entrypoint": "daymine"
}
EOF

cp "${repo_root}/README.md" "${binary_dir}/README.md"
cp "${repo_root}/LICENSE" "${binary_dir}/LICENSE"
cp "${dist_dir}/release-manifest.json" "${binary_dir}/release-manifest.json"

tar -czf "${dist_dir}/${artifact_name}" \
  -C "${repo_root}" \
  "dist/daymine-${goos}-${goarch}" \
  docs/generated/panel-skills.md \
  docs/product-specs/self-hosted-agent-dashboard.md

echo "${dist_dir}/${artifact_name}"
