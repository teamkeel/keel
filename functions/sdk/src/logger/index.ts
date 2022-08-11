import chalk, { ChalkInstance } from 'chalk';

export enum Level {
  Info = 'info',
  Error = 'error',
  Debug = 'debug',
  Warn = 'warn'
}

export interface LoggerOptions {
  transport?: Transport
  colorize?: boolean
  timestamps?: boolean
}

type LevelColors = Record<Level, ChalkInstance>

const LevelColorPalette : LevelColors = {
  'error': chalk.red,
  'info': chalk.cyan,
  'debug': chalk.magenta,
  'warn': chalk.yellow
};

export interface Transport {
  log: (msg: Msg, level: Level, options: LoggerOptions) => void
}

// The default (and only) transport implementation of Logger class
// logs to STDOUT / STDERR
export class ConsoleTransport implements Transport {
  log = (msg: Msg, level: Level, options: LoggerOptions) : void => {
    if (options.timestamps) {
      const dateFormatOpts: Intl.DateTimeFormatOptions = {
        year: 'numeric', month: '2-digit', day: '2-digit',
        hour: 'numeric', minute: 'numeric', second: 'numeric',
        hour12: false
      };

      msg = `[${new Date().toLocaleDateString('en-GB', dateFormatOpts)}] ${msg}`;
    }

    if (options.colorize) {
      const color = LevelColorPalette[level];

      console[level](color(msg));
    } else {
      console[level](msg);
    }
  };
}

type Msg = string | boolean | undefined | null | number

// Usage: new Logger({ colorize: true }).log('foo', Level.Info);
export default class Logger {
  private readonly options : LoggerOptions = {
    colorize: true,
    transport: new ConsoleTransport(),
    timestamps: true
  };

  constructor(opts?: LoggerOptions) {
    if (opts) {
      this.options = {
        ...this.options,
        ...opts
      };
    }
  }

  log = (msg: Msg, level: Level = Level.Info) : void => {
    this.options.transport.log(msg, level, this.options);
  };
}
