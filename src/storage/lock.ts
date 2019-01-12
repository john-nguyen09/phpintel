import { EventEmitter } from "events";

export class Lock {
    private _locked = false;
    private _ee = new EventEmitter();

    async acquire() {
        return new Promise<void>(resolve => {
            if (!this._locked) {
                this._locked = true;
                return resolve();
            }

            const tryAcquire = () => {
                if (!this._locked) {
                    this._locked = true;
                    this._ee.removeListener('release', tryAcquire);
                    return resolve();
                }
            };

            this._ee.on('release', tryAcquire);
        });
    }

    release() {
        this._locked = false;
        setImmediate(() => this._ee.emit('release'));
    }
}