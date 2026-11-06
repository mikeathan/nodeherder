/*
 * Spatial heat/display types shared by the canvas and the recorded-positions view. Session-only
 * frame analytics were removed with the browser recorder; positions now come from the hub.
 */

/** Aggregated target-seconds for one heat cell. */
export type HeatCell = { x: number; y: number; count: number; intensity: number };
/** Milliseconds per second for sampling and display conversions. */
export const MILLISECONDS_PER_SECOND = 1000;
/** Heatmap grid size, in millimetres. */
export const HEAT_CELL_MM = 250;
