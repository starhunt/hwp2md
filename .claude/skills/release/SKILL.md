---
name: release
description: Automates versioning and release notes generation for hwp2markdown. Use this skill when preparing a new release, creating version tags, or generating changelogs from git history.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
user-invocable: true
---

# Release Automation Skill

이 스킬은 hwp2markdown의 버저닝과 릴리즈 노트 생성을 자동화합니다.

## 사용법

```
/release [version] [options]
```

**예시:**
- `/release` - 다음 버전 자동 결정 및 릴리즈 준비
- `/release v0.2.0` - 특정 버전으로 릴리즈
- `/release patch` - 패치 버전 증가 (v0.1.0 → v0.1.1)
- `/release minor` - 마이너 버전 증가 (v0.1.0 → v0.2.0)
- `/release major` - 메이저 버전 증가 (v0.1.0 → v1.0.0)

## 릴리즈 프로세스

### 1. 버전 결정

현재 태그 확인:
```bash
git tag -l 'v*' --sort=-version:refname | head -5
```

버전 형식: `vMAJOR.MINOR.PATCH` (Semantic Versioning)

**버전 증가 기준:**
- **MAJOR**: 하위 호환성이 깨지는 변경
- **MINOR**: 새로운 기능 추가 (하위 호환)
- **PATCH**: 버그 수정, 문서 수정

### 2. 변경사항 분석

마지막 태그 이후 커밋 분석:
```bash
# 마지막 태그 찾기
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

# 커밋 로그 (태그 이후 또는 전체)
if [ -n "$LAST_TAG" ]; then
  git log $LAST_TAG..HEAD --oneline
else
  git log --oneline
fi
```

### 3. CHANGELOG.md 생성/업데이트

CHANGELOG 형식 (Keep a Changelog):

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2024-01-15

### Added
- 새로운 기능 설명

### Changed
- 변경된 기능 설명

### Fixed
- 버그 수정 설명

### Removed
- 제거된 기능 설명

## [0.1.0] - 2024-01-01

### Added
- Initial release
```

**커밋 메시지 → 카테고리 매핑:**
| 커밋 접두사 | 카테고리 |
|-------------|----------|
| `Add`, `Implement`, `Create` | Added |
| `Update`, `Change`, `Refactor`, `Improve` | Changed |
| `Fix`, `Resolve`, `Correct` | Fixed |
| `Remove`, `Delete`, `Drop` | Removed |
| `Deprecate` | Deprecated |
| `Security` | Security |

### 4. 릴리즈 태그 생성

```bash
# 태그 생성 (annotated tag)
git tag -a v0.2.0 -m "Release v0.2.0

주요 변경사항:
- 기능 1
- 기능 2
- 버그 수정
"

# 원격에 푸시
git push origin v0.2.0
```

### 5. GitHub Release (선택)

goreleaser를 사용한 릴리즈:
```bash
# 로컬 테스트 (실제 릴리즈 없이)
goreleaser release --snapshot --clean

# 실제 릴리즈 (CI에서 자동 실행)
goreleaser release --clean
```

## 릴리즈 체크리스트

릴리즈 전 확인사항:

- [ ] 모든 테스트 통과 (`make test`)
- [ ] 린트 검사 통과 (`make lint`)
- [ ] CHANGELOG.md 업데이트
- [ ] README.md 버전 정보 확인
- [ ] main 브랜치에 모든 변경사항 병합
- [ ] 이전 릴리즈 이후 breaking change 확인

## 자동 릴리즈 노트 생성

커밋 히스토리 기반 릴리즈 노트:

```bash
# 커밋을 카테고리별로 분류
git log $LAST_TAG..HEAD --pretty=format:"%s" | while read msg; do
  case "$msg" in
    Add*|Implement*|Create*) echo "### Added"; echo "- $msg" ;;
    Update*|Change*|Refactor*) echo "### Changed"; echo "- $msg" ;;
    Fix*) echo "### Fixed"; echo "- $msg" ;;
    Remove*|Delete*) echo "### Removed"; echo "- $msg" ;;
  esac
done
```

## 예시: 전체 릴리즈 워크플로우

```bash
# 1. 현재 상태 확인
git status
make test

# 2. 마지막 태그 확인
git describe --tags --abbrev=0

# 3. 변경사항 확인
git log $(git describe --tags --abbrev=0)..HEAD --oneline

# 4. CHANGELOG 업데이트 (수동 또는 자동)

# 5. 변경사항 커밋
git add CHANGELOG.md
git commit -m "docs: Update CHANGELOG for v0.2.0"

# 6. 태그 생성 및 푸시
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin main
git push origin v0.2.0

# 7. GitHub Actions가 자동으로 goreleaser 실행
```

## 긴급 패치 릴리즈

핫픽스가 필요한 경우:

```bash
# 1. 핫픽스 브랜치 생성
git checkout -b hotfix/v0.1.1 v0.1.0

# 2. 수정 적용
# ... 코드 수정 ...

# 3. 커밋 및 태그
git commit -m "Fix: Critical bug description"
git tag -a v0.1.1 -m "Hotfix release v0.1.1"

# 4. main에 병합 및 푸시
git checkout main
git merge hotfix/v0.1.1
git push origin main
git push origin v0.1.1
```
