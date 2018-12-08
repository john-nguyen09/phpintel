import { Class } from "../symbol/class/class";
import { TypeName } from "../type/name";
import { Serializer } from "./serializer";
import { Symbol } from "../symbol/symbol";
import * as fs from 'fs';
import * as path from 'path';
import { Function } from "../symbol/function/function";
import { Variable } from "../symbol/variable/variable";
import { Parameter } from "../symbol/variable/parameter";
import { Constant } from "../symbol/constant/constant";
import { ClassConstant } from "../symbol/constant/classConstant";
import { Union, Literal } from "runtypes";
import { Method } from "../symbol/function/method";
import { Property } from "../symbol/variable/property";

interface SymbolEncoding {
    encode(symbol: Symbol): Buffer;
    decode(buffer: Buffer): Symbol;
}

const classEncoding =  {
    encode: (symbol: Class): Buffer => {
        let serializer = new Serializer();

        serializer.writeTypeName(symbol.name);
        serializer.writeTypeName(symbol.extend);
        serializer.writeLocation(symbol.location);
        serializer.writeSymbolModifier(symbol.modifier);

        serializer.writeInt32(symbol.implements.length);
        for (let name of symbol.implements) {
            serializer.writeTypeName(name);
        }

        serializer.writeInt32(symbol.traits.length);
        for (let name of symbol.traits) {
            serializer.writeTypeName(name);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): Class => {
        let serializer = new Serializer(buffer);
        let theClass = new Class();

        theClass.name = serializer.readTypeName() || new TypeName('');
        theClass.extend = serializer.readTypeName();
        theClass.location = serializer.readLocation();
        theClass.modifier = serializer.readSymbolModifier();
        
        let noImplements = serializer.readInt32();
        for (let i = 0; i < noImplements; i++) {
            theClass.implements.push(serializer.readTypeName() || new TypeName(''));
        }
        
        let noTraits = serializer.readInt32();
        for (let i = 0; i < noTraits; i++) {
            theClass.traits.push(serializer.readTypeName() || new TypeName(''));
        }

        return theClass;
    }
}

const functionEncoding = {
    encode: (symbol: Function): Buffer => {
        let serializer = new Serializer();

        serializer.writeTypeName(symbol.name);
        serializer.writeLocation(symbol.location);
        serializer.writeString(symbol.description);

        serializer.writeInt32(symbol.types.length);
        for (let type of symbol.types) {
            serializer.writeTypeName(type);
        }

        serializer.writeInt32(Object.keys(symbol.scopeVar.variables).length);
        for (let varName in symbol.scopeVar.variables) {
            let variable = symbol.scopeVar.variables[varName];

            serializer.writeString(variable.name);
            serializer.writeTypeComposite(variable.type);
        }

        serializer.writeInt32(symbol.parameters.length);
        for (let param of symbol.parameters) {
            serializer.writeString(param.name);
            serializer.writeTypeComposite(param.type);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): Function => {
        let func = new Function();
        let serializer = new Serializer(buffer);

        func.name = serializer.readTypeName() || new TypeName('');
        func.location = serializer.readLocation();
        func.description = serializer.readString();

        let noTypes = serializer.readInt32();
        for (let i = 0; i < noTypes; i++) {
            func.typeAggregate.push(serializer.readTypeName() || new TypeName(''));
        }

        let noVariables = serializer.readInt32();
        for (let i = 0; i < noVariables; i++) {
            func.scopeVar.set(new Variable(serializer.readString(), serializer.readTypeComposite()));
        }

        let noParams = serializer.readInt32();
        for (let i = 0; i < noParams; i++) {
            let param = new Parameter();

            param.name = serializer.readString();
            param.type = serializer.readTypeComposite();
            func.parameters.push(param);
        }

        return func;
    }
};

const constEncoding = {
    encode: (symbol: Constant): Buffer => {
        let serializer = new Serializer();

        serializer.writeTypeName(symbol.name);
        serializer.writeLocation(symbol.location);
        serializer.writeTypeName(symbol.type);
        serializer.writeString(symbol.value);

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): Constant => {
        let constant = new Constant();
        let serializer = new Serializer(buffer);

        constant.name = serializer.readTypeName() || new TypeName('');
        constant.location = serializer.readLocation();
        constant.resolvedType = serializer.readTypeName();
        constant.resolvedValue = serializer.readString();

        return constant;
    }
};

const classConstEncoding = {
    encode: (symbol: ClassConstant): Buffer => {
        let serializer = new Serializer();

        serializer.writeTypeName(symbol.name);
        serializer.writeLocation(symbol.location);
        serializer.writeTypeName(symbol.type);
        serializer.writeTypeName(symbol.scope);
        serializer.writeString(symbol.value);

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): ClassConstant => {
        let classConst = new ClassConstant();
        let serializer = new Serializer(buffer);

        classConst.name = serializer.readTypeName() || new TypeName('');
        classConst.location = serializer.readLocation();
        classConst.type = serializer.readTypeName() || new TypeName('');
        classConst.scope = serializer.readTypeName();
        classConst.value = serializer.readString();

        return classConst;
    }
};

const methodEncoding = {
    encode: (symbol: Method): Buffer => {
        let serializer = new Serializer();

        serializer.writeTypeName(symbol.name);
        serializer.writeLocation(symbol.location);
        serializer.writeString(symbol.description);
        serializer.writeSymbolModifier(symbol.modifier);
        serializer.writeTypeName(symbol.scope);

        serializer.writeInt32(symbol.types.length);
        for (let type of symbol.types) {
            serializer.writeTypeName(type);
        }

        serializer.writeInt32(Object.keys(symbol.variables).length);
        for (let varName in symbol.variables) {
            serializer.writeString(varName);
            serializer.writeTypeComposite(symbol.variables[varName].type);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): Method => {
        let method = new Method();
        let serializer = new Serializer(buffer);

        method.name = serializer.readTypeName() || new TypeName('');
        method.location = serializer.readLocation();
        method.description = serializer.readString();
        method.modifier = serializer.readSymbolModifier();
        method.scope = serializer.readTypeName();

        let noTypes = serializer.readInt32();
        for (let i = 0; i < noTypes; i++) {
            method.pushType(serializer.readTypeName());
        }

        let noVariables = serializer.readInt32();
        for (let i = 0; i < noVariables; i++) {
            method.setVariable(new Variable(serializer.readString(), serializer.readTypeComposite()));
        }

        return method;
    }
};

const propertyEncoding = {
    encode: (symbol: Property): Buffer => {
        let serializer = new Serializer();

        serializer.writeString(symbol.name);
        serializer.writeLocation(symbol.location);
        serializer.writeString(symbol.description);
        serializer.writeSymbolModifier(symbol.modifier);
        serializer.writeTypeName(symbol.scope);
        serializer.writeTypeComposite(symbol.type);

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): Property => {
        let property = new Property();
        let serializer = new Serializer(buffer);

        property.name = serializer.readString();
        property.location = serializer.readLocation();
        property.description = serializer.readString();
        property.modifier = serializer.readSymbolModifier();
        property.scope = serializer.readTypeName();
        property.type = serializer.readTypeComposite();

        return property;
    }
};

enum Type {
    Class = 1,
    Function = 2,
    Constant = 3,
    ClassConstant = 4,
    Method = 5,
    Property = 6,
};

const symbolEncodingMap: {[key: number]: SymbolEncoding} = {
    1: classEncoding,
    2: functionEncoding,
    3: constEncoding,
    4: classConstEncoding,
    5: methodEncoding,
    6: propertyEncoding,
};

export = {
    type: 'symbol-encoding',
    encode: (symbol: Symbol): Buffer => {
        let serializer = new Serializer();
        let symbolType: Type | null = null;

        if (symbol instanceof Class) {
            symbolType = Type.Class;
        } else if (symbol instanceof Function) {
            symbolType = Type.Function;
        } else if (symbol instanceof Constant) {
            symbolType = Type.Constant;
        } else if (symbol instanceof ClassConstant) {
            symbolType = Type.ClassConstant;
        } else if (symbol instanceof Method) {
            symbolType = Type.Method;
        } else if (symbol instanceof Property) {
            symbolType = Type.Property;
        }

        if (symbolType !== null) {
            serializer.writeInt32(symbolType);

            let encoding = symbolEncodingMap[symbolType];
            serializer.writeBuffer(encoding.encode(symbol));
        }

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): Symbol => {
        if (buffer.byteLength == 0) {
            throw new Error(`Invalid buffer`);            
        }

        let serializer = new Serializer(buffer);
        let type = serializer.readInt32();
        buffer = serializer.readBuffer();

        if (type in symbolEncodingMap) {
            return symbolEncodingMap[type].decode(buffer);
        }

        throw new Error(`Invalid buffer`);
    },
    buffer: false
} as Level.Encoding;
