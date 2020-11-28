export class Logger {
  private _scope: string = 'main'

  constructor(scope?: string) {
    if (scope) {
      this._scope = scope
    }
  }

  public error(error: Error) {
    console.error('[%cNEKO%c] [' + this._scope + '] %cERR', 'color: #498ad8;', '', 'color: #d84949;', error)
  }

  public warn(...log: any[]) {
    console.warn('[%cNEKO%c] [' + this._scope + '] %cWRN', 'color: #498ad8;', '', 'color: #eae364;', ...log)
  }

  public info(...log: any[]) {
    console.info('[%cNEKO%c] [' + this._scope + '] %cINF', 'color: #498ad8;', '', 'color: #4ac94c;', ...log)
  }

  public debug(...log: any[]) {
    console.log('[%cNEKO%c] [' + this._scope + '] %cDBG', 'color: #498ad8;', '', 'color: #eae364;', ...log)
  }
}
