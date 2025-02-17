const parseInterval = require("postgres-interval");

const isoRegex =
  /^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$/;

class Duration {
  constructor(postgresString) {
    this._typename = "Duration";
    this.pgInterval = postgresString;
    if (!this.pgInterval) {
      this._interval = parseInterval(postgresString);
    }
  }

  static fromISOString(isoString) {
    // todo parse iso string to postgres string
    const match = isoString.match(isoRegex);
    if (match) {
      let d = new Duration();
      d._interval.years = match[1];
      d._interval.months = match[2];
      d._interval.days = match[3];
      d._interval.hours = match[4];
      d._interval.minutes = match[5];
      d._interval.seconds = match[6];
      return d;
    }
    return new Duration();
  }

  toISOString() {
    return this._interval.toISOStringShort();
  }

  toPostgres() {
    return this._interval.toPostgres();
  }
}

module.exports = {
  Duration,
};
