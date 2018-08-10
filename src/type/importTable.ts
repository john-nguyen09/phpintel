import { TypeName } from './name';
import { NamespaceName } from '../symbol/name/namespaceName';

export class ImportTable {
    private namespace: NamespaceName;
    private imports: {[key: string]: string} = {};

    constructor() { }

    public setNamespace(namespace: NamespaceName) {
        this.namespace = namespace;
    }

    public import(fqn: string, alias?: string) {
        let parts = fqn.split('\\').filter((part) => {
            return part != '';
        });

        if (alias == undefined) {
            alias = parts.pop();
        }

        if (alias != undefined) {
            this.imports[alias] = '\\' + parts.join('\\');
        }
    }

    public getFqn(name: string) {
        if (TypeName.isFullyQualifiedName(name)) {
            return name;
        }

        let parts = name.split('\\');
        let alias = parts[0];
        let namespace: string;

        if (alias != undefined && alias in this.imports) {
            namespace = this.imports[alias];
        } else {
            namespace = this.namespace != null ? this.namespace.fqn : '';
        }

        return namespace + (parts.length > 0 ? '\\' + parts.join('\\') : '');
    }
}
