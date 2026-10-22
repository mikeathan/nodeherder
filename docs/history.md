# Device history

Open a device's History tab or choose **View full history** from an entity dialog. The dialog
link keeps the selected entity, ahead of a saved preference. History stays device-scoped; mobile
navigation offers Home, Devices, Automations, Assistant and More (the remaining destinations).

Choose an entity chip and a period, including Today, Yesterday or a Custom window of at most
31 elapsed days. Each device remembers its period and entity in browser storage; Custom also
remembers its validated bounds. Invalid saved bounds fall back to 24h. Calendar boundaries and
labels follow the browser's named time zone, including daylight-saving changes.

Numeric charts support mouse drag to select a range, wheel/pinch zoom, and a one-finger readout.
The buttons provide zoom, pan and reset. Focus the chart and use `+`, `-`, arrows and `0` for the
same operations. Zoom only changes the visible portion of the loaded window; it sends no new
history query or device command. Missing values remain gaps; the end marker is an actual sample.

**Statistics** is available for numeric readings except energy counters. Daily bars and their
accessible table report sample minimum, average, maximum, last and count. Statistics use every
loaded raw sample, converted to the displayed unit before aggregation. Edge days are labelled
partial days. Binary entities keep state durations; events retain repeated reports; energy keeps
its observed-consumption summary.

**Export CSV** uses the selected period, including data outside a zoomed view. History exports
raw `timestamp,value,unit`; Statistics exports
`bucket_start,bucket_end,min,average,max,last,samples,unit` in the displayed unit. ISO timestamps
use UTC (`Z`); missing values have empty cells. Spreadsheet formula-like text is escaped.

A capped response labels the actual covered sample window and warns that newer samples may be
missing. Export is disabled. **Narrow this range** changes to a Custom window only when selected.
Loading, empty and failure messages remain explicit; refreshing retries a read, never a command.

The phone sheet has Close and Escape fallbacks. A downward drag greater than 120px on its header
dismisses it when content is at the top and reduced motion is off. Its controls and mobile
navigation respect safe-area insets. The web manifest is optional; no offline service worker is
installed by this change.
