import { Symbol, Consumer } from "../symbol";
import { NamespaceUseClause } from "./useClause";
import { NamespaceName } from "../name/namespaceName";
import { Alias } from "./alias";

export class NamespaceUse extends Symbol implements Consumer {
    private baseNamespace: string = '';
    private useClauses: NamespaceUseClause[] = [];

    consume(other: Symbol): boolean {
        if (other instanceof NamespaceUseClause) {
            this.useClauses.push(other);

            return true;
        } else if (other instanceof NamespaceName) {
            this.baseNamespace = other.fqn;

            return true;
        }

        return false;
    }

    get aliasTable(): Alias[] {
        let results: Alias[] = [];

        for (let useClause of this.useClauses) {
            let alias = useClause.alias;
            let fqn = this.baseNamespace + '\\' + useClause.namespaceName;

            results.push(new Alias(alias, fqn));
        }

        return results;
    }
}