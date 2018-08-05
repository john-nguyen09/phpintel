import { Symbol, Consumer } from "../symbol";
import { NamespaceAliasClause } from "./aliasClause";
import { NamespaceName } from "../name/namespaceName";

export class NamespaceUseClause extends Symbol implements Consumer {
    private _namepsaceName: string = '';
    private _alias: string = '';
    
    consume(other: Symbol): boolean {
        if (other instanceof NamespaceAliasClause) {
            this._alias = other.name;

            return true;
        } else if (other instanceof NamespaceName) {
            this._namepsaceName = other.name;

            return true;
        }

        return false;
    }

    get alias(): string {
        if (this._alias != null) {
            return this._alias;
        }

        return this._namepsaceName.substr(this._namepsaceName.lastIndexOf('\\') + 1);
    }

    get namespaceName(): string {
        return this._namepsaceName;
    }
}