import { TypeName } from "./name";
import { nonenumerable } from "../util/decorator";

export class TypeComposite {
    @nonenumerable
    protected existingTypes: { [key: string]: boolean } = {};
    protected _types: TypeName[] = [];

    push(type: TypeName) {
        if (type == undefined || type.name in this.existingTypes) {
            return;
        }

        this._types.push(type);
        this.existingTypes[type.name] = true;
    }

    clone(): TypeComposite {
        let result = new TypeComposite();

        for (let type of this.types) {
            result.push(type);
        }

        return result;
    }

    get types(): TypeName[] {
        return this._types;
    }
}