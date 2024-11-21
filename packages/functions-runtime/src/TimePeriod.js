class TimePeriod {
  constructor(period = "", value = 0, offset = 0, complete = false) {
    this.period = period;
    this.value = value;
    this.offset = offset;
    this.complete = complete;
  }

  static fromExpression(expression) {
    // Regex pattern
    const pattern =
      /^(this|next|last)?\s*(\d+)?\s*(complete)?\s*(second|minute|hour|day|week|month|year|seconds|minutes|hours|days|weeks|months|years)?$/i;

    const shorthandPattern = /^(now|today|tomorrow|yesterday)$/i;

    const shorthandMatch = shorthandPattern.exec(expression.trim());
    if (shorthandMatch) {
      const shorthand = shorthandMatch[1].toLowerCase();
      switch (shorthand) {
        case "now":
          return new TimePeriod();
        case "today":
          return TimePeriod.fromExpression("this day");
        case "tomorrow":
          return TimePeriod.fromExpression("next complete day");
        case "yesterday":
          return TimePeriod.fromExpression("last complete day");
      }
    }

    const match = pattern.exec(expression.trim());
    if (!match) {
      throw new Error("Invalid time period expression");
    }

    const [, direction, rawValue, isComplete, rawPeriod] = match;

    let period = rawPeriod ? rawPeriod.toLowerCase().replace(/s$/, "") : "";
    let value = rawValue ? parseInt(rawValue, 10) : 1;
    let complete = Boolean(isComplete);
    let offset = 0;

    switch (direction?.toLowerCase()) {
      case "this":
        offset = 0;
        complete = true;
        break;
      case "next":
        offset = complete ? 1 : 0;
        break;
      case "last":
        offset = -value;
        break;
      default:
        throw new Error(
          "Time period expression must start with this, next, or last"
        );
    }

    return new TimePeriod(period, value, offset, complete);
  }

  periodStartSQL() {
    let sql = "NOW()";
    if (this.offset !== 0) {
      sql = `${sql} + INTERVAL '${this.offset} ${this.period}'`;
    }

    if (this.complete) {
      sql = `DATE_TRUNC('${this.period}', ${sql})`;
    } else {
      sql = `(${sql})`;
    }

    return sql;
  }

  periodEndSQL() {
    let sql = this.periodStartSQL();
    if (this.value != 0) {
      sql = `(${sql} + INTERVAL '${this.value} ${this.period}')`;
    }
    return sql;
  }
}

module.exports = {
  TimePeriod,
  File,
};
