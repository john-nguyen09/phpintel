import { Reference } from "../symbol/reference";
import { App } from "../app";
import { FunctionTable } from "../storage/table/function";
import { TypeComposite } from "../type/composite";
import { Function } from "../symbol/function/function";
import { PhpDocument } from "../symbol/phpDocument";

export namespace RefResolver {
    export async function getFuncSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Function[]> {
        const funcTable = App.get<FunctionTable>(FunctionTable);
        
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveToFullyQualified(phpDoc.importTable);

        return await funcTable.get(ref.type.getName());
    }
}