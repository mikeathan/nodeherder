/*
 * Rigid two-dimensional calibration between measured radar positions and known room
 * positions. Both frames are millimetres, so only rotation and translation are solved;
 * nothing is scaled, clamped or applied implicitly.
 */

export const CALIBRATION_MIN_PAIRS = 2;
export const CALIBRATION_MAX_PAIRS = 5;
/** RMS above which the result is usable only with a visible warning. */
export const CALIBRATION_WARN_MM = 200;
/** RMS above which the pairs disagree too much to apply. */
export const CALIBRATION_BLOCK_MM = 500;
const DEGREES_TO_RADIANS = Math.PI / 180;
const RADIANS_TO_DEGREES = 180 / Math.PI;

export type CalibrationPair = {
  /** Live slot the measured position came from; null when the pair is incomplete. */
  slot: number | null;
  measuredX: number;
  measuredY: number;
  roomX: number;
  roomY: number;
};

export type CalibrationSolution = {
  /** Degrees, normalised to (-180, 180]. */
  rotation: number;
  /** Room-frame translation in millimetres. */
  x: number;
  y: number;
  /** Root-mean-square residual over all pairs. */
  rms: number;
  /** Largest individual pair residual. */
  worst: number;
  residuals: number[];
};

/** Per-pair distances after applying a solution, in input order. */
export function calibrationResiduals(
  pairs: readonly CalibrationPair[],
  solution: Pick<CalibrationSolution, 'rotation' | 'x' | 'y'>
): number[] {
  const radians = solution.rotation * DEGREES_TO_RADIANS;
  const cos = Math.cos(radians);
  const sin = Math.sin(radians);
  return pairs.map((pair) => {
    const x = cos * pair.measuredX - sin * pair.measuredY + solution.x;
    const y = sin * pair.measuredX + cos * pair.measuredY + solution.y;
    return Math.hypot(x - pair.roomX, y - pair.roomY);
  });
}

/** Normalises an angle to (-180, 180]. */
export function normalizeDegrees(value: number): number {
  const wrapped = ((((value + 180) % 360) + 360) % 360) - 180;
  return wrapped === -180 ? 180 : wrapped;
}

/**
 * Solves the rigid transform `room = R(rotation)·measured + translation` by least squares.
 * It returns a blocking error when the pairs cannot determine a pose or disagree too much.
 */
export function solveCalibration(pairs: readonly CalibrationPair[]): {
  error: string;
  warning: string;
  solution: CalibrationSolution | null;
} {
  if (pairs.length < CALIBRATION_MIN_PAIRS || pairs.length > CALIBRATION_MAX_PAIRS) {
    return { error: `Add between ${CALIBRATION_MIN_PAIRS} and ${CALIBRATION_MAX_PAIRS} reference pairs.`, warning: '', solution: null };
  }
  for (const pair of pairs) {
    if (![pair.measuredX, pair.measuredY, pair.roomX, pair.roomY].every(Number.isFinite)) {
      return { error: 'Every pair needs finite measured and room coordinates.', warning: '', solution: null };
    }
  }
  const spread = Math.max(
    ...pairs.flatMap((left, index) =>
      pairs.slice(index + 1).map((right) => Math.hypot(left.measuredX - right.measuredX, left.measuredY - right.measuredY))
    )
  );
  if (!(spread > 1)) {
    return { error: 'Measured positions collapse to one point; move the target between captures.', warning: '', solution: null };
  }

  const count = pairs.length;
  const measuredX = pairs.reduce((sum, pair) => sum + pair.measuredX, 0) / count;
  const measuredY = pairs.reduce((sum, pair) => sum + pair.measuredY, 0) / count;
  const roomX = pairs.reduce((sum, pair) => sum + pair.roomX, 0) / count;
  const roomY = pairs.reduce((sum, pair) => sum + pair.roomY, 0) / count;
  let dot = 0;
  let cross = 0;
  for (const pair of pairs) {
    const mx = pair.measuredX - measuredX;
    const my = pair.measuredY - measuredY;
    const rx = pair.roomX - roomX;
    const ry = pair.roomY - roomY;
    dot += mx * rx + my * ry;
    cross += mx * ry - my * rx;
  }
  if (!Number.isFinite(dot) || !Number.isFinite(cross) || (Math.abs(dot) < 1e-9 && Math.abs(cross) < 1e-9)) {
    return { error: 'These pairs cannot determine a rotation; choose points further apart.', warning: '', solution: null };
  }
  const rotation = normalizeDegrees(Math.atan2(cross, dot) * RADIANS_TO_DEGREES);
  const radians = rotation * DEGREES_TO_RADIANS;
  const x = roomX - (Math.cos(radians) * measuredX - Math.sin(radians) * measuredY);
  const y = roomY - (Math.sin(radians) * measuredX + Math.cos(radians) * measuredY);
  if (![rotation, x, y].every(Number.isFinite)) {
    return { error: 'The solution is not finite; check the entered coordinates.', warning: '', solution: null };
  }
  const residuals = calibrationResiduals(pairs, { rotation, x, y });
  const rms = Math.sqrt(residuals.reduce((sum, value) => sum + value * value, 0) / residuals.length);
  const worst = Math.max(...residuals);
  if (rms > CALIBRATION_BLOCK_MM) {
    return {
      error: `The reference points disagree by ${Math.round(rms)} mm RMS (limit ${CALIBRATION_BLOCK_MM} mm). Recheck each room position.`,
      warning: '',
      solution: null,
    };
  }
  return {
    error: '',
    warning:
      rms > CALIBRATION_WARN_MM
        ? `Residual is ${Math.round(rms)} mm RMS; the result is usable but check the room positions.`
        : '',
    solution: { rotation, x, y, rms, worst, residuals },
  };
}
