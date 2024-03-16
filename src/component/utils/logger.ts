export class Logger {
  // eslint-disable-next-line
  constructor(
    protected readonly _scope: string = 'main',
    private readonly _color: boolean = false // !!process.env.VUE_APP_LOG_COLOR, // TODO: add support for color
  ) {}

  protected _console(level: string, m: string, fields?: Record<string, any>) {
    const scope = this._scope

    let t = ''
    const args = []
    for (const name in fields) {
      if (fields[name] instanceof Error) {
        if (this._color) {
          t += ' %c%s="%s"%c'
          args.push('color:#d84949;', name, (fields[name] as Error).message, '')
        } else {
          t += ' %s="%s"'
          args.push(name, (fields[name] as Error).message)
        }
        continue
      }

      if (typeof fields[name] === 'string' || fields[name] instanceof String) {
        t += this._color ? ' %c%s=%c"%s"' : ' %s="%s"'
      } else {
        t += this._color ? ' %c%s=%c%o' : ' %s=%o'
      }

      if (this._color) {
        args.push('color:#498ad8;', name, '', fields[name])
      } else {
        args.push(name, fields[name])
      }
    }

    if (this._color) {
      switch (level) {
        case 'error':
          console.error('[%cNEKO%c] [%s] %cERR%c %s' + t, 'color:#498ad8;', '', scope, 'color:#d84949;', '', m, ...args)
          break
        case 'warn':
          console.warn('[%cNEKO%c] [%s] %cWRN%c %s' + t, 'color:#498ad8;', '', scope, 'color:#eae364;', '', m, ...args)
          break
        case 'info':
          console.info('[%cNEKO%c] [%s] %cINF%c %s' + t, 'color:#498ad8;', '', scope, 'color:#4ac94c;', '', m, ...args)
          break
        default:
        case 'debug':
          console.debug('[%cNEKO%c] [%s] %cDBG%c %s' + t, 'color:#498ad8;', '', scope, 'color:#eae364;', '', m, ...args)
          break
      }
    } else {
      switch (level) {
        case 'error':
          console.error('[NEKO] [%s] ERR %s' + t, scope, m, ...args)
          break
        case 'warn':
          console.warn('[NEKO] [%s] WRN %s' + t, scope, m, ...args)
          break
        case 'info':
          console.info('[NEKO] [%s] INF %s' + t, scope, m, ...args)
          break
        default:
        case 'debug':
          console.debug('[NEKO] [%s] DBG %s' + t, scope, m, ...args)
          break
      }
    }
  }

  public error(message: string, fields?: Record<string, any>) {
    this._console('error', message, fields)
  }

  public warn(message: string, fields?: Record<string, any>) {
    this._console('warn', message, fields)
  }

  public info(message: string, fields?: Record<string, any>) {
    this._console('info', message, fields)
  }

  public debug(message: string, fields?: Record<string, any>) {
    this._console('debug', message, fields)
  }
}
