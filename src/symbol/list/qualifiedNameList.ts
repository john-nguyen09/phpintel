import { Symbol } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";

export class QualifiedNameList extends Symbol {
    public symbols: QualifiedName[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.symbols.push(other);

            return true;
        }

        return false;
    }

    get names(): string[] {
        let names: string[] = [];

        for (let symbol of this.symbols) {
            names.push(symbol.name);
        }

        return names;
    }
}