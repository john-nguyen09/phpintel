import { TypeName } from './name';
import { NamespaceName } from '../symbol/name/namespaceName';

export class ImportTable {
    public namespace: NamespaceName = new NamespaceName();
    public imports: {[key: string]: string} = {};

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
        if (TypeName.isBuiltin(name) || TypeName.isFullyQualifiedName(name)) {
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

    public getQualified(fqn: string) {
        if (TypeName.isBuiltin(fqn) || !TypeName.isFullyQualifiedName(fqn)) {
            return fqn;
        }

        let parts = fqn.split('\\');
        parts.shift();
        let index = parts.length - 1;

        for (; index >= 0; index--) {
            if (parts[index] in this.imports) {
                break;
            }
        }

        const isRoot = typeof this.namespace === 'undefined' || this.namespace.isRoot;

        // No part is found in import table
        if (!isRoot && index < 0) {
            return fqn;
        }

        const qualifiedName = parts.slice(index).join('\\');

        if (isRoot) {
            return qualifiedName;
        }

        return '\\' + qualifiedName;
    }
}
