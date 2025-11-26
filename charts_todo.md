# Metrics & Visualization Improvements

## 📊 Current System Analysis

### Backend Metrics Collection
- ✅ Data stored in BoltDB with time-series structure
- ✅ Metrics collected per device expose (temperature, humidity, etc.)
- ✅ Configurable retention period (keep for X days)
- ⚠️ No data sampling for large datasets yet
- ⚠️ Binary data (presence, contact, state ON/OFF) treated same as numeric

### Frontend Visualization
- ⚠️ Single chart type for all data types (not optimized for binary)
- ⚠️ Color distribution needs improvement
- ⚠️ No preview charts in entity dialogs
- ⚠️ Charts not user-friendly for quick insights

---

## 🎯 Phase 1: Binary Data Visualization (Session 1)

### Backend Changes
- [ ] Add data type classification system
  - [ ] Identify binary exposes (boolean, ON/OFF, open/close, present/absent)
  - [ ] Identify numeric exposes (temperature, humidity, brightness)
  - [ ] Identify enum exposes (color, mode)
- [ ] Create binary data aggregation endpoint
  - [ ] Return state duration (how long in each state)
  - [ ] Return state change events with timestamps
  - [ ] Return percentage time in each state
- [ ] Optimize binary data queries
  - [ ] Add indexing for faster binary data retrieval
  - [ ] Sample binary data intelligently (state changes only)

### Frontend Changes
- [ ] Create binary data chart component
  - [ ] Timeline view showing ON/OFF periods as colored blocks
  - [ ] Bar chart showing time percentage in each state
  - [ ] Event markers for state changes
- [ ] Update metrics viewer to detect data type
  - [ ] Route binary data to binary chart component
  - [ ] Route numeric data to line chart component
- [ ] Add binary data color scheme
  - [ ] Consistent colors: ON=green, OFF=gray, OPEN=red, CLOSED=green
  - [ ] Present=blue, Absent=gray

**Acceptance Criteria:**
- Binary sensors show timeline blocks instead of line charts
- Clear visual indication of state durations
- Smooth transitions between states

---

## 🎨 Phase 2: Chart Color & UX Improvements (Session 2)

### Color System
- [ ] Create color palette system
  - [ ] Temperature: red (hot) to blue (cold) gradient
  - [ ] Humidity: blue gradient
  - [ ] Battery: green (high) to red (low)
  - [ ] Power/Energy: orange/yellow
  - [ ] Light levels: yellow gradient
  - [ ] CO2/Air Quality: green (good) to red (poor)
- [ ] Add theme support (light/dark mode)
- [ ] Make colors accessible (WCAG compliant)

### Chart Improvements
- [ ] Add chart type selection
  - [ ] Line chart (default for continuous data)
  - [ ] Area chart (for cumulative data like energy)
  - [ ] Bar chart (for discrete measurements)
  - [ ] Step chart (for values that change instantly)
- [ ] Add chart controls
  - [ ] Zoom in/out
  - [ ] Pan through time
  - [ ] Reset view
  - [ ] Toggle data series on/off
- [ ] Improve tooltips
  - [ ] Show exact timestamp
  - [ ] Show all series values at that point
  - [ ] Show duration since last change (for binary)
  - [ ] Format values with units

### User Experience
- [ ] Add loading states and skeletons
- [ ] Add empty state messages
- [ ] Add error handling with retry
- [ ] Optimize chart rendering performance
- [ ] Add responsive design for mobile

**Acceptance Criteria:**
- Charts use intuitive, consistent colors
- Users can interact with charts (zoom, pan)
- Tooltips provide useful context
- Charts load quickly even with large datasets

---

## 📈 Phase 3: Mini Preview Charts (Session 3)

### Entity Dialog Preview
- [ ] Design mini chart component (Home Assistant style)
  - [ ] Small sparkline for numeric data (last 24h)
  - [ ] Mini timeline for binary data (last 24h)
  - [ ] Show current value prominently
  - [ ] Show min/max/avg for numeric data
- [ ] Add preview to device entity cards
  - [ ] Fetch last 24h of data on card open
  - [ ] Cache preview data to avoid repeated requests
  - [ ] Update preview when new data arrives
- [ ] Add "View Full History" button
  - [ ] Navigate to full metrics page
  - [ ] Pre-select the entity

### Device Card Enhancements
- [ ] Show multiple expose previews per device
- [ ] Add quick stats (last update, battery level, etc.)
- [ ] Color-code cards by status (online/offline, battery low)

**Acceptance Criteria:**
- Each entity shows a small preview chart
- Preview updates in real-time (optional)
- Quick access to detailed metrics
- Minimal performance impact

---

## 🔧 Phase 4: Advanced Features (Session 4)

### Data Aggregation
- [ ] Implement data sampling strategies
  - [ ] Min/Max/Avg bucketing for large datasets
  - [ ] Keep raw data for recent period (last 24h)
  - [ ] Downsample older data (daily averages)
- [ ] Add comparison features
  - [ ] Compare same metric across devices
  - [ ] Compare different time periods
  - [ ] Show year-over-year trends

### Export & Analysis
- [ ] Add data export functionality
  - [ ] CSV export
  - [ ] JSON export
  - [ ] Date range selection
- [ ] Add statistical analysis
  - [ ] Moving averages
  - [ ] Trend lines
  - [ ] Anomaly detection highlights

### Smart Insights
- [ ] Add automatic insights
  - [ ] "Temperature increased 5°C in last hour"
  - [ ] "Door opened 15 times today"
  - [ ] "Average humidity: 65% (normal)"
- [ ] Add threshold visualization
  - [ ] Show user-defined thresholds on charts
  - [ ] Highlight threshold violations
  - [ ] Link to automation triggers

**Acceptance Criteria:**
- Large datasets load quickly with sampled data
- Users can export data for external analysis
- Charts provide intelligent insights
- Threshold violations are visually obvious

---

## 🏗️ Phase 5: Architecture & Performance (Session 5)

### Backend Optimization
- [ ] Implement proper indexing in BoltDB
  - [ ] Index by device ID + expose name + timestamp
  - [ ] Optimize range queries
- [ ] Add caching layer
  - [ ] Cache common queries (last 24h, last 7d)
  - [ ] Invalidate cache on new data
  - [ ] Set appropriate TTLs
- [ ] Add query pagination
  - [ ] Limit data points returned per request
  - [ ] Support cursor-based pagination
- [ ] Add WebSocket streaming for real-time updates
  - [ ] Stream new data points to open charts
  - [ ] Reduce polling frequency

### Frontend Optimization
- [ ] Implement virtual scrolling for data tables
- [ ] Add chart lazy loading
  - [ ] Load charts when visible (intersection observer)
  - [ ] Unload off-screen charts
- [ ] Optimize re-renders
  - [ ] Memoize expensive computations
  - [ ] Use Web Workers for data processing
- [ ] Add progressive data loading
  - [ ] Show low-resolution data first
  - [ ] Load high-resolution on zoom

### Testing
- [ ] Create metrics mock data generator
  - [ ] Generate realistic time-series data
  - [ ] Generate binary state changes
  - [ ] Generate edge cases (gaps, outliers)
- [ ] Add performance benchmarks
  - [ ] Measure query performance
  - [ ] Measure chart render times
  - [ ] Set performance budgets

**Acceptance Criteria:**
- Queries return in < 100ms for common ranges
- Charts render smoothly (60fps)
- Real-time updates work without polling
- System handles 1000+ devices efficiently

---

## 📋 Implementation Checklist

### Pre-Implementation
- [ ] Review current BoltDB schema
- [ ] Audit all expose types across devices
- [ ] Create design mockups for new chart types
- [ ] Set up metrics testing environment

### Documentation
- [ ] Document data type classification system
- [ ] Document color palette and usage
- [ ] Create chart component API documentation
- [ ] Add user guide for metrics features

### Migration
- [ ] Create data migration plan if schema changes
- [ ] Test with production data sample
- [ ] Plan rollback strategy

---

## 🎯 Success Metrics

### User Experience
- Binary data is immediately understandable
- Charts load in < 2 seconds
- Users can find insights within 5 clicks
- Mobile experience is smooth

### Technical
- 90% reduction in data transferred for large ranges
- Cache hit rate > 80%
- Zero N+1 queries
- < 50ms backend query time (p95)

### Feature Adoption
- 80% of users view metrics weekly
- 50% of users use preview charts
- 30% of users export data

---

## 🔄 Iteration Plan

1. **Week 1-2**: Phase 1 - Binary Data Visualization
2. **Week 3-4**: Phase 2 - Chart Colors & UX
3. **Week 5**: Phase 3 - Mini Preview Charts
4. **Week 6-7**: Phase 4 - Advanced Features
5. **Week 8**: Phase 5 - Performance & Polish

Each phase should be:
- Independently deployable
- User-tested before moving to next phase
- Documented with examples