import { TypeName } from "./name";

export class TypeComposite {
    protected _types: TypeName[] = [];

    push(type: TypeName) {
        this._types.push(type);
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