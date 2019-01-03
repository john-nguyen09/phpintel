import { TypeName } from "./name";

export class TypeComposite {
    protected _types: TypeName[] = [];

    push(type: TypeName) {
        if (type == undefined) {
            return;
        }

        this._types.push(type);
    }

    clone(): TypeComposite {
        let result = new TypeComposite();

        for (let type of this.types) {
            result.push(type);
        }

        return result;
    }

    get types(): TypeName[] {
        let result: TypeName[] = [];

        for (let type of this._types) {
            let doesContain = false;

            for (let currType of result)  {
                if (type.isSameAs(currType)) {
                    doesContain = true;

                    break;
                }
            }

            if (!doesContain) {
                result.push(type);
            }
        }

        return result;
    }
}