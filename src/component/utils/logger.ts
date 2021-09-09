export class Logger {
  protected _scope: string = 'main'

  constructor(scope?: string) {
    if (scope) {
      this._scope = scope
    }
  }

  protected _console(level: string, m: string, fields?: Record<string, any>) {
    let t = ''
    const args = []
    for (const name in fields) {
      if (typeof fields[name] === 'string' || fields[name] instanceof String) {
        t += ' %c%s=%c"%s"'
      } else {
        t += ' %c%s=%c%o'
      }

      args.push('color:#498ad8;', name, '', fields[name])
    }

    const scope = this._scope
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
