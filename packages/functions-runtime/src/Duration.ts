import parseInterval from "postgres-interval";

const isoRegex =
  /^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$/;

interface Interval {
  years?: number;
  months?: number;
  days?: number;
  hours?: number;
  minutes?: number;
  seconds?: number;
  toISOStringShort(): string;
  toPostgres(): string;
}

export class Duration {
  private _typename: string;
  private pgInterval: string;
  private _interval: Interval;

  constructor(postgresString: string) {
    this._typename = "Duration";
    this.pgInterval = postgresString;
    this._interval = parseInterval(postgresString);
  }

  static fromISOString(isoString: string): Duration {
    const match = isoString.match(isoRegex);
    if (match) {
      const d = new Duration("0");
      d._interval.years = match[1] ? parseInt(match[1]) : undefined;
      d._interval.months = match[2] ? parseInt(match[2]) : undefined;
      d._interval.days = match[3] ? parseInt(match[3]) : undefined;
      d._interval.hours = match[4] ? parseInt(match[4]) : undefined;
      d._interval.minutes = match[5] ? parseInt(match[5]) : undefined;
      d._interval.seconds = match[6] ? parseInt(match[6]) : undefined;
      return d;
    }
    return new Duration("0");
  }

  toISOString(): string {
    return this._interval.toISOStringShort();
  }

  toPostgres(): string {
    return this._interval.toPostgres();
  }
}
