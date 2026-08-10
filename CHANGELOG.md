# SendLog Syslog Changelog

## [1.7.0] - 2026-08-10

### ✨ New Features
- feat(ui): expose message format selector for raw-pri mode (989e430)
- feat(syslog): add raw-pri message format for vendor log replay (dc227f8)
- feat: add CONTRIBUTING.md and SECURITY.md for improved guidelines and security practices (4908713)

### 🐛 Bug Fixes
- fix(release): generate checksums for DEB and AppImage packages (078b2f2)
- fix(lint): check ignored errors and drop dead test scaffolding (47e304b)

### 📚 Documentation
- docs: update CHANGELOG for v1.6.0 (efc7765)

### 🔧 Maintenance
- chore(deps): bump the dev-tools group across 1 directory with 2 updates (ca3734d)
- chore(deps): bump react-hook-form from 7.71.2 to 7.72.0 in /frontend (2279371)
- chore(deps): bump tailwindcss in /frontend in the ui group (6ea5bc5)
- chore(deps): bump the nextjs group in /frontend with 2 updates (5c356ef)
- chore(ci): bump pnpm/action-setup from 4 to 5 (2dcff5f)
- chore(deps): bump @types/node from 25.4.0 to 25.5.0 in /frontend (0124b9c)

### 📝 Other Changes
- ci: unblock test, lint, frontend and build-check jobs (10c9457)



## [1.6.0] - 2026-03-10

### ✨ New Features
- feat: implement batch import functionality for log files with support for CSV and JSON formats (7b426bb)
- feat: add connection profiles and log templates management with API integration (0c8560e)
- feat: implement TemplateManager component for managing log templates (c8ff62b)
- feat: add profile manager component for saving and loading connection profiles (b8b50fe)
- feat: implement continuous/stress test functionality with UI components and backend integration (42d11a8)
- feat: add /docs to .gitignore to exclude documentation files from version control (57b70ea)

### ♻️ Refactoring
- refactor(frontend): split form.tsx into modular components following SOLID principles (cb3fa4d)

### 📚 Documentation
- docs: update CHANGELOG for v1.5.0 (5829743)

### 🔧 Maintenance
- chore: update Go version to 1.25 in release workflow (97e6a04)
- chore(ci): bump actions/setup-node from 4 to 6 (2d08759)
- chore(ci): bump actions/setup-go from 5 to 6 (21b9176)
- chore(ci): bump actions/download-artifact from 4 to 8 (f96fbc8)

### 📝 Other Changes
- Refactor application structure to implement service-oriented architecture (f6b440b)



## [1.5.0] - 2026-03-10

### ✨ New Features
- feat: update changelog and package version to 1.5.0 with UI/UX enhancements and technical improvements (b556360)
- feat: enhance layout and styling for improved user experience (6fd3e4f)
- feat: enhance application functionality with new URL handling and exit confirmation (1a2b758)
- feat: add support for custom CA and client certificates for mTLS in connection handling (6bfee71)
- feat: enhance changelog update process in release workflow (5078458)

### 📚 Documentation
- docs: update README for improved clarity and structure, enhance feature descriptions (ce2c85f)

### 📝 Other Changes
- Update dependencies and Go version in go.mod and go.sum (e74258b)



## [1.5.0] - 2026-03-09

### 🎨 UI/UX Overhaul - Fluent Design
- **Fluent Design System**: Complete UI redesign following Microsoft WinUI 3 guidelines
  - Modern card-based layout with subtle shadows and rounded corners
  - Consistent 36px (h-9) input heights across all form elements
  - Improved spacing and visual hierarchy
- **Custom App Icon**: New professional icon with Fluent Design aesthetics
  - Blue gradient background (#0078D4 to #005A9E)
  - Document with log lines and circular send arrow
  - Multiple sizes generated (16px to 1024px)
  - Properly embedded icon for all platforms
- **Fixed Window Size**: Disabled resize/maximize for consistent experience
  - 900x700 fixed dimensions
  - Clean, focused interface without window chrome distractions

### ⚡ Technical Improvements
- **Tailwind CSS v4**: Migrated to latest Tailwind with CSS-first configuration
  - Uses `@import "tailwindcss"` and `@theme` block syntax
  - Removed deprecated tailwind.config.ts
  - PostCSS integration with `@tailwindcss/postcss`
- **Next.js 16**: Updated to latest Next.js with Turbopack
- **Form Alignment**: Fixed grid layout for Connection Settings
  - 12-column responsive grid system
  - Proper label-input spacing with space-y-1
  - Button alignment with pb-[3px] fine-tuning

### 🖥️ Platform Improvements
- **Linux**: webkit2gtk-4.1 support via `-tags webkit2_41`
- **Windows**: Mica backdrop effect with transparent webview
- **macOS**: Native title bar styling with dark appearance

### 📦 Dependencies Updated
- React 19.2.4
- Zod 4.3.6 (with new validation API)
- Lucide React 0.577.0
- All Radix UI components updated to latest

---

## [1.3.0] - 2025-11-17

### 🏗️ Architecture Improvements
- **Professional RFC 6587 Implementation**: Complete refactoring of TCP framing module
  - Created dedicated `framing.go` module with clean separation of concerns
  - Eliminated code spaghetti with modular, testable architecture
  - Industry-standard implementation following RFC 6587 specifications

### ⚡ Performance Optimizations
- **Efficient Memory Management**:
  - Pre-calculated buffer sizes to avoid memory reallocations
  - `bytes.Buffer` usage for optimal memory construction
  - Batch processing capability with `FrameBatch()` method
- **Benchmark Results**:
  - Octet Counting: 204.4 ns/op, 98 B/op, 2 allocs/op
  - Non-Transparent: 107.1 ns/op, 96 B/op, 1 allocs/op
  - Batch Processing: Highly efficient for multiple messages

### ✅ Comprehensive Validation
- **UTF-8 Validation**: RFC 5424 compliance with `utf8.ValidString()`
- **Message Length Limits**: Configurable maximum message length
- **Framing-Specific Validation**:
  - Non-Transparent: Rejects messages containing LF characters
  - Octet Counting: No content restrictions (as per RFC 6587)
- **Empty Message Detection**: Prevents invalid empty messages

### 🧪 Complete Test Coverage
- **Unit Tests**:
  - `TestOctetCountingFraming`: Verifies RFC 6587 Section 3.4.1 compliance
  - `TestNonTransparentFraming`: Verifies RFC 6587 Section 3.4.2 compliance
  - `TestFramingValidation`: Edge cases and validation rules
  - `TestFrameBatch`: Batch processing functionality
- **Benchmarks**:
  - `BenchmarkOctetCountingFraming`
  - `BenchmarkNonTransparentFraming`
  - `BenchmarkFrameBatch`
- **All Tests Passing**: 100% test success rate

### 📚 Enhanced Documentation
- **Technical Documentation**: New `TECHNICAL_DOCUMENTATION.md` with:
  - Complete RFC 6587 implementation details
  - Architecture diagrams and data flow
  - Best practices and optimization techniques
  - Performance benchmarks and analysis
  - Integration examples (Backend & Frontend)
- **Inline Code Documentation**:
  - Detailed comments with RFC references
  - Implementation examples in comments
  - Clear explanation of limitations and trade-offs

### 🔧 Backend Improvements
- **Framer Component**:
  ```go
  type Framer struct {
      config FramingConfig
  }
  ```
  - `Frame(message)`: Single message framing with validation
  - `FrameBatch(messages)`: Efficient batch processing
  - `NewFramer(config)`: Constructor with flexible configuration
- **Configuration Options**:
  ```go
  type FramingConfig struct {
      Method           FramingMethod
      ValidateUTF8     bool
      MaxMessageLength int
  }
  ```
- **Helper Functions**:
  - `IsValidFramingMethod()`: Method validation
  - `RecommendedFramingMethod()`: Returns RFC-recommended method
  - `DefaultFramingConfig()`: Sensible defaults

### 🎯 Enhanced Error Handling
- **Descriptive Error Messages**: Context-rich error information
- **RFC References in Errors**: Error messages cite relevant RFC sections
- **Validation Before Processing**: Fail-fast approach prevents invalid operations
- **Error Propagation**: Clear error chains from validation to UI

### 🔄 Backward Compatibility
- **Maintained API Compatibility**: No breaking changes to existing functionality
- **TLS Support Preserved**: All TLS features remain functional
- **RFC Compliance**: Both RFC 5424 and RFC 3164 continue to work
- **UI Unchanged**: Frontend maintains same user experience

### 📦 Code Quality
- **No Code Spaghetti**: Clean, modular design with single responsibility principle
- **DRY Principle**: Eliminated code duplication through abstraction
- **Type Safety**: Strong typing throughout with proper Go idioms
- **Memory Efficiency**: Zero unnecessary allocations in hot paths

### 🚀 Standards Compliance
- **RFC 6587**: Full implementation of TCP transmission standard
  - Section 3.4.1: Octet Counting (recommended method)
  - Section 3.4.2: Non-Transparent Framing (legacy support)
- **RFC 5424**: UTF-8 validation and message format compliance
- **RFC 5425**: TLS/SSL transport (from v1.2.0, maintained)
- **RFC 3164**: BSD syslog protocol support (maintained)

### 🎨 UI Improvements
- **Framing Method Selector**:
  - Clear labels: "Octet Counting (RFC 6587)" with star indicator
  - "Non-Transparent (LF)" for legacy systems
  - Disabled when UDP is selected (TCP-only feature)
  - Auto-selects recommended method (Octet Counting)

### 🔬 Developer Experience
- **Easy Testing**: `go test -v` runs all tests
- **Performance Profiling**: `go test -bench=. -benchmem` for benchmarks
- **Clean Module Structure**: Easy to understand and maintain
- **Comprehensive Examples**: Test files serve as usage examples

### 🐛 Bug Fixes
- **Fixed Buffer Sizing**: Correct pre-calculation prevents overflows
- **UTF-8 Handling**: Proper validation of multi-byte characters
- **LF Detection**: Accurate detection in non-transparent framing
- **Message Length**: Correct byte counting for all character encodings

## [1.2.0] - 2025-11-17

### 🔒 Security Features
- **TLS/SSL Support**: Added secure syslog transmission over TLS (RFC 5425)
  - Supports TLS 1.2 and TLS 1.3 for modern security standards
  - Configurable certificate verification (enable/disable)
  - Accept self-signed certificates for development/testing environments
  - Auto-suggest port 6514 when TLS is enabled
  - Connection state logging with TLS version and cipher suite information
- **Enhanced Connection Protocol**: 
  - TCP connections now support TLS encryption
  - UDP remains unencrypted (as per protocol limitations)
  - Smart connection handling with proper timeout management (10s)
  - Explicit TLS handshake verification

### ✨ UI Enhancements
- **TLS Configuration Controls**:
  - "Use TLS/SSL" checkbox for enabling encrypted connections
  - "Verify Certificate" checkbox for certificate validation control
  - Auto-disable TLS options when UDP is selected
  - Smart port switching between 514 (plain) and 6514 (TLS)
  - Enhanced connection status messages with security information
  - Visual feedback for self-signed certificate connections

### 🔧 Technical Improvements
- **Backend (Go)**:
  - New `dialTLS()` function with professional TLS configuration
  - Unified `dialConnection()` function for protocol abstraction
  - Updated `CheckConnection()` to support TLS parameters
  - Updated `SendSyslogMessages()` with TLS capability
  - Proper error handling for TLS handshake failures
  - Connection state management for secure connections
- **Frontend (TypeScript)**:
  - Updated form schema with `UseTLS` and `TLSVerify` fields
  - Auto-regenerated Wails bindings for type safety
  - Enhanced connection toggle with TLS status display
  - Improved toast notifications with security indicators

### 📦 Dependencies
- Uses Go standard library `crypto/tls` (no external dependencies)
- Maintains backward compatibility with non-TLS connections

### 🐛 Bug Fixes
- Proper connection cleanup on TLS handshake failures
- Prevented TLS usage with UDP protocol
- Consistent error messaging across connection types

### 📚 Documentation
- Updated README.md with TLS/SSL feature documentation
- Added security notes for certificate verification
- Updated usage instructions with TLS configuration steps
- Added standard port information (514 vs 6514)

## [1.1.0] - 2025-11-17

### 🎨 UI/UX Improvements
- **Desktop-Optimized Layout**: Fixed window size (900x780) for optimal desktop experience without scrollbars
- **Modern Card-Based Design**: Reorganized UI with shadcn/ui Card components for better organization
- **Improved Spacing**: Reduced padding and margins throughout for better content density
- **Compact Components**: Optimized input heights (h-9), button sizes, and font sizes for desktop
- **Custom Scrollbar**: Added thin, styled scrollbar with better aesthetics
- **Icon Integration**: Added lucide-react icons (CheckCircle, XCircle, Send, Network) for better visual feedback

### ✨ Form Enhancements
- **Structured Layout**: Reorganized into three main cards (Connection Settings, Message Configuration, Send Messages)
- **Better Visual Hierarchy**: Clear sections with icons and descriptive titles
- **Responsive Grid**: 2-column layout for form fields to maximize space efficiency
- **Connection Status**: Visual connect/disconnect button with icons and color coding
- **Improved Select Menus**: Fixed height selects (h-9) with scrollable content and max-height constraints
- **Compact Textarea**: Optimized message input area with fixed 100px height and no resize
- **Better Labels**: Shorter, more concise form labels and descriptions

### 🔧 Technical Improvements
- **Window Configuration**: 
  - Width: 900px (min: 900, max: 1100)
  - Height: 780px (min: 780, max: 900)
  - Disabled unwanted resizing that caused scroll issues
- **Overflow Control**: Proper scroll management with custom scrollbar classes
- **Toast Notifications**: Cleaner, more informative feedback with checkmarks and error symbols
- **Code Organization**: Better component structure, cleaner imports, modern React patterns
- **Layout Optimization**: Header height reduced, better flex layout for content area

### 📦 Dependencies
- Added lucide-react for modern icon set
- Enhanced shadcn/ui components usage (Card, CardHeader, CardContent, CardTitle)
- Updated component styling with Tailwind utility classes
- Maintained full compatibility with existing Wails backend

### 🐛 Bug Fixes
- Fixed layout scroll issues on desktop application
- Resolved component sizing inconsistencies across different sections
- Improved form validation display with smaller error messages
- Fixed toast notifications overflow and improved readability
- Corrected spacing issues in header and main content area

### 🗑️ Removed
- Removed max-w-4xl constraint from main form container
- Removed unnecessary padding from layout sections
- Removed redundant FormDescription elements for cleaner UI
- Removed min-h-screen class causing scroll issues

## [1.0.0] - 2023-10-01

### Added
- **Dual Protocol Support**: Send log data over both TCP and UDP.
- **Simple Setup**: Easy to install and configure for quick integration.
- **Reliable (TCP)**: Ensures reliable delivery of log data with automatic retries.
- **Fast and Lightweight (UDP)**: For scenarios where speed is prioritized over reliability.
- **RFC 5424 & RFC 3164**: Support for both modern and legacy syslog formats
- **TCP Framing**: Octet counting and non-transparent framing methods

### Fixed
- Initial release, no fixes yet.

### Changed
- Initial release, no changes yet.