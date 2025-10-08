# GitHub Issues for goimgserver Implementation

This directory contains GitHub issues broken down into manageable, self-contained tasks for implementing the goimgserver image processing service using **Test-Driven Development (TDD)**.

## TDD Implementation Approach

**MANDATORY**: All tasks must follow the Red-Green-Refactor TDD cycle:

1. **🔴 RED**: Write failing tests first that define the desired behavior
2. **🟢 GREEN**: Write minimal code to make the tests pass
3. **🔵 REFACTOR**: Improve code quality while keeping tests green

### TDD Success Criteria for All Issues:
- [ ] All tests written before implementation code
- [ ] No production code without corresponding tests
- [ ] Test coverage >95% for each component
- [ ] Tests cover error scenarios and edge cases
- [ ] Implementation is minimal and focused on making tests pass

## Implementation Order

### Phase 1: Foundation (High Priority)
- **gh-0002.md** - Core Application Configuration and Command-Line Arguments
- **gh-0003.md** - Image Processing Engine with bimg Integration
- **gh-0015.md** - File Resolution System Implementation *(NEW)*
- **gh-0004.md** - Cache Management System
- **gh-0005.md** - Image Endpoint Implementation
- **gh-0014.md** - Default Image Fallback System

### Phase 2: Administrative Features (Medium Priority)
- **gh-0006.md** - Command Endpoints Implementation
- **gh-0007.md** - Pre-cache Initialization System
- **gh-0008.md** - HTTP Server Enhancement and Middleware
- **gh-0009.md** - Comprehensive Error Handling and Logging
- **gh-0013.md** - Security Hardening and Production Safety

### Phase 3: Quality & Production (Medium to Low Priority)
- **gh-0010.md** - Testing Suite Implementation
- **gh-0011.md** - Documentation and Deployment Guide
- **gh-0012.md** - Performance Optimization and Monitoring

## Dependencies

```
gh-0002 (Config) 
├── gh-0003 (Image Processing)
├── gh-0008 (HTTP Server)
│
gh-0003 + gh-0015 (File Resolution)
├── gh-0004 (Cache) 
├── gh-0005 (Image Endpoints)
├── gh-0007 (Pre-cache)
├── gh-0014 (Default Image System)
│
gh-0002 + gh-0004
├── gh-0006 (Command Endpoints)
│
All Core (gh-0002 to gh-0007, gh-0015)
├── gh-0009 (Error Handling)
├── gh-0010 (Testing)
├── gh-0011 (Documentation)
├── gh-0012 (Performance)
```

## Notes

- **gh-0015.md** is the new file resolution system supporting extension auto-detection and grouped images *(NEW)*
- **gh-0014.md** is the default image fallback system (high priority for user experience)
- Each task is designed to be self-contained with clear acceptance criteria
- Dependencies are clearly marked to ensure proper implementation order
- High priority tasks (gh-0002, gh-0003, gh-0015, gh-0004, gh-0005, gh-0014) form the core functionality
- Medium priority tasks enhance the service for production use
- Low priority tasks focus on optimization and documentation

## New Features in This Update

### File Resolution System (gh-0015)
- **Extension Auto-Detection**: `GET /img/cat` automatically finds `cat.jpg`, `cat.png`, or `cat.webp`
- **Extension Priority**: jpg/jpeg > png > webp when multiple extensions exist
- **Grouped Images**: Support for organized images in folders
  - `GET /img/cats` → serves `cats/default.*`
  - `GET /img/cats/fluffy` → serves `cats/fluffy.*`
- **Zero 404 Errors**: Combined with default image system for complete fallback coverage

### Updated Design Documents
All design documents have been restructured with number prefixes for better organization:
- `00-overview.md` - Main project overview
- `01-tdd-methodology.md` - TDD implementation guidelines  
- `02-endpoints.md` - API endpoint specifications
- `03-url-parsing.md` - URL parsing strategy
- `04-file-resolution.md` - File resolution system *(NEW)*
- `05-default-image.md` - Default image system
- `06-api-specification.md` - Complete API specification
- `07-security.md` - Security considerations
- `08-deployment.md` - Deployment guide

## Estimated Timeline

- **Phase 1**: 5-6 weeks (core functionality with file resolution, comprehensive TDD + default image system)
- **Phase 2**: 3-4 weeks (administrative features + security with TDD)
- **Phase 3**: 2-3 weeks (quality and production readiness with TDD)

Total estimated time: 10-13 weeks depending on team size and TDD experience.

**Note**: TDD adds initial development time but significantly reduces debugging, maintenance, and refactoring time. The new file resolution system and default image system add significant user experience value with flexible image access patterns.