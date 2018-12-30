import { Reference, RefKind } from "../symbol/reference";
import { App } from "../app";
import { FunctionTable } from "../storage/table/function";
import { TypeComposite } from "../type/composite";
import { Function } from "../symbol/function/function";
import { PhpDocument } from "../symbol/phpDocument";
import { Class } from "../symbol/class/class";
import { ClassTable } from "../storage/table/class";
import { Method } from "../symbol/function/method";
import { MethodTable } from "../storage/table/method";
import { Property } from "../symbol/variable/property";
import { PropertyTable } from "../storage/table/property";
import { ClassConstant } from "../symbol/constant/classConstant";
import { ClassConstantTable } from "../storage/table/classConstant";

export namespace RefResolver {
    export async function getFuncSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Function[]> {
        const funcTable = App.get<FunctionTable>(FunctionTable);
        
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveToFullyQualified(phpDoc.importTable);

        return await funcTable.get(ref.type.getName());
    }

    export async function getMethodSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Method[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        const methodTable = App.get<MethodTable>(MethodTable);
        let className: string = '';
        let methodName: string = '';

        if (ref.refKind === RefKind.ClassTypeDesignator) {
            ref.type.resolveToFullyQualified(phpDoc.importTable);
            className = ref.type.getName();
            methodName = '__construct';
        } else if (ref.refKind === RefKind.Method && ref.scope !== null) {
            ref.scope.resolveToFullyQualified(phpDoc.importTable);
            className = ref.scope.getName();
            methodName = ref.type.getName();
        }

        return await methodTable.searchByClass(className, methodName);
    }

    export async function getPropSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Property[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        const propTable = App.get<PropertyTable>(PropertyTable);
        let className = '';

        if (ref.scope !== null) {
            ref.scope.resolveToFullyQualified(phpDoc.importTable);
            className = ref.scope.getName();
        }

        return await propTable.searchByClass(className, ref.type.getName());
    }

    export async function getClassSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Class[]> {
        const classTable = App.get<ClassTable>(ClassTable);

        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveToFullyQualified(phpDoc.importTable);

        return await classTable.get(ref.type.getName());
    }

    export async function getClassConstSymbols(phpDoc: PhpDocument, ref: Reference): Promise<ClassConstant[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        const classConstTable = App.get<ClassConstantTable>(ClassConstantTable);
        let className = '';

        if (ref.scope !== null) {
            ref.scope.resolveToFullyQualified(phpDoc.importTable);
            className = ref.scope.getName();
        }

        return await classConstTable.searchByClass(className, ref.type.getName());
    }
}