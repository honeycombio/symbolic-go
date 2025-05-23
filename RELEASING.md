# Releasing

- `Add steps to prepare release`
- Update `CHANGELOG.md` with the changes since the last release.
- Commit changes, push, and open a release preparation pull request for review.
- Once the pull request is merged, fetch the updated `main` branch.
- Create a new branch named `release/vX.Y.Z` (fill in X.Y.Z with your version number) and push that.
  - CircleCI will build the symbolic libraries + C dependencies and commit those to your branch
- Fetch your branch with CircleCI's commit and apply a tag for the new version (e.g. `git tag -a v1.2.3 -m "v1.2.3"`)
- Push the tag upstream e.g. `git push origin v1.2.3`
- Copy changelog entry for newest version into draft GitHub release created as part of CI publish steps.
  - Make sure to "generate release notes" in github for full changelog notes and any new contributors
- Publish the github draft release and this will kick off publishing to GitHub and the NPM registry.