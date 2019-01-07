import { Symbol, Consumer } from "../symbol";
import { NamespaceName } from "../name/namespaceName";

export class NamespaceDefinition extends Symbol implements Consumer {
    public name: NamespaceName;

    consume(other: Symbol): boolean {
        if (other instanceof NamespaceName) {
            this.name = other;
        }

        return true;
    }
}