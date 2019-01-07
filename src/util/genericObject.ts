export function createObject(constructor: Function): Object {
    let GenericObject = (() => {
        function Constructor() { }

        if (constructor != undefined) {
            Constructor.prototype.constructor = constructor;
        }

        return Constructor as any;
    })();

    return new GenericObject();
}