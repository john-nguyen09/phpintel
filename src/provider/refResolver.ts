import { Reference } from "../symbol/reference";
import { App } from "../app";
import { FunctionTable } from "../storage/table/function";
import { TypeComposite } from "../type/composite";
import { Function } from "../symbol/function/function";
import { PhpDocument } from "../symbol/phpDocument";
import { Class } from "../symbol/class/class";
import { ClassTable } from "../storage/table/class";
import { Method } from "../symbol/function/method";
import { MethodTable } from "../storage/table/method";

export namespace RefResolver {
    export async function getFuncSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Function[]> {
        const funcTable = App.get<FunctionTable>(FunctionTable);
        
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveToFullyQualified(phpDoc.importTable);

        return await funcTable.get(ref.type.getName());
    }

    export async function getClassConstructorSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Method[]> {
        const methodTable = App.get<MethodTable>(MethodTable);

        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveToFullyQualified(phpDoc.importTable);

        return await methodTable.searchByClass(ref.type.getName(), '__construct');
    }

    export async function getClassSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Class[]> {
        const classTable = App.get<ClassTable>(ClassTable);

        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveToFullyQualified(phpDoc.importTable);

        return await classTable.get(ref.type.getName());
    }
}