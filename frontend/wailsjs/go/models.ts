export namespace logger {
	
	export class LogEntry {
	    id: string;
	    timestamp: number;
	    level: string;
	    message: string;
	    accountId?: string;
	    providerId?: string;
	    requestId?: string;
	    data?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.message = source["message"];
	        this.accountId = source["accountId"];
	        this.providerId = source["providerId"];
	        this.requestId = source["requestId"];
	        this.data = source["data"];
	    }
	}
	export class LogOptions {
	    level: string;
	    limit: number;
	    offset: number;
	    keyword: string;
	
	    static createFrom(source: any = {}) {
	        return new LogOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.limit = source["limit"];
	        this.offset = source["offset"];
	        this.keyword = source["keyword"];
	    }
	}

}

export namespace types {
	
	export class Account {
	    id: string;
	    providerId: string;
	    name: string;
	    email?: string;
	    credentials: Record<string, string>;
	    status: string;
	    lastUsed?: number;
	    createdAt: number;
	    updatedAt: number;
	    errorMessage?: string;
	    requestCount?: number;
	    dailyLimit?: number;
	    todayUsed?: number;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.providerId = source["providerId"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.credentials = source["credentials"];
	        this.status = source["status"];
	        this.lastUsed = source["lastUsed"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.errorMessage = source["errorMessage"];
	        this.requestCount = source["requestCount"];
	        this.dailyLimit = source["dailyLimit"];
	        this.todayUsed = source["todayUsed"];
	    }
	}
	export class ApiKey {
	    id: string;
	    key: string;
	    name: string;
	    enabled: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    expiresAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new ApiKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.expiresAt = this.convertValues(source["expiresAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ManagementApiConfig {
	    enabled: boolean;
	    port: number;
	    host: string;
	    authToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new ManagementApiConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.port = source["port"];
	        this.host = source["host"];
	        this.authToken = source["authToken"];
	    }
	}
	export class LoadBalanceConfig {
	    strategy: string;
	    excludeFailed: boolean;
	    recoveryTime: number;
	
	    static createFrom(source: any = {}) {
	        return new LoadBalanceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategy = source["strategy"];
	        this.excludeFailed = source["excludeFailed"];
	        this.recoveryTime = source["recoveryTime"];
	    }
	}
	export class ContextManagementConfig {
	    enabled: boolean;
	    maxContextMessages: number;
	    summaryModel?: string;
	    summaryThreshold: number;
	    enableAutoSummary: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ContextManagementConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.maxContextMessages = source["maxContextMessages"];
	        this.summaryModel = source["summaryModel"];
	        this.summaryThreshold = source["summaryThreshold"];
	        this.enableAutoSummary = source["enableAutoSummary"];
	    }
	}
	export class ToolCallingConfig {
	    enabled: boolean;
	    forceAdapter?: string;
	    allowedProtocols?: string[];
	    disabledProtocols?: string[];
	    customPromptTemplate?: string;
	    diagnosticsEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolCallingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.forceAdapter = source["forceAdapter"];
	        this.allowedProtocols = source["allowedProtocols"];
	        this.disabledProtocols = source["disabledProtocols"];
	        this.customPromptTemplate = source["customPromptTemplate"];
	        this.diagnosticsEnabled = source["diagnosticsEnabled"];
	    }
	}
	export class SessionConfig {
	    enabled: boolean;
	    maxHistoryLength: number;
	    deleteAfterChat: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SessionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.maxHistoryLength = source["maxHistoryLength"];
	        this.deleteAfterChat = source["deleteAfterChat"];
	    }
	}
	export class ProxyConfig {
	    enabled: boolean;
	    port: number;
	    host: string;
	    enableApiKey: boolean;
	    apiKeys?: ApiKey[];
	    sessionConfig: SessionConfig;
	    toolCallingConfig: ToolCallingConfig;
	    contextConfig: ContextManagementConfig;
	    loadBalanceConfig: LoadBalanceConfig;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.port = source["port"];
	        this.host = source["host"];
	        this.enableApiKey = source["enableApiKey"];
	        this.apiKeys = this.convertValues(source["apiKeys"], ApiKey);
	        this.sessionConfig = this.convertValues(source["sessionConfig"], SessionConfig);
	        this.toolCallingConfig = this.convertValues(source["toolCallingConfig"], ToolCallingConfig);
	        this.contextConfig = this.convertValues(source["contextConfig"], ContextManagementConfig);
	        this.loadBalanceConfig = this.convertValues(source["loadBalanceConfig"], LoadBalanceConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppConfig {
	    proxyConfig: ProxyConfig;
	    managementConfig: ManagementApiConfig;
	    theme: string;
	    startMinimized: boolean;
	    autoStart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyConfig = this.convertValues(source["proxyConfig"], ProxyConfig);
	        this.managementConfig = this.convertValues(source["managementConfig"], ManagementApiConfig);
	        this.theme = source["theme"];
	        this.startMinimized = source["startMinimized"];
	        this.autoStart = source["autoStart"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ToolFunction {
	    name: string;
	    arguments: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolFunction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.arguments = source["arguments"];
	    }
	}
	export class ToolCall {
	    id: string;
	    type: string;
	    function: ToolFunction;
	
	    static createFrom(source: any = {}) {
	        return new ToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.function = this.convertValues(source["function"], ToolFunction);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChatMessage {
	    role: string;
	    content: string;
	    name?: string;
	    tool_call_id?: string;
	    tool_calls?: ToolCall[];
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.name = source["name"];
	        this.tool_call_id = source["tool_call_id"];
	        this.tool_calls = this.convertValues(source["tool_calls"], ToolCall);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class ModelMappingEntry {
	    clientModel: string;
	    providerId: string;
	    accountId: string;
	    providerModel: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelMappingEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clientModel = source["clientModel"];
	        this.providerId = source["providerId"];
	        this.accountId = source["accountId"];
	        this.providerModel = source["providerModel"];
	        this.enabled = source["enabled"];
	    }
	}
	export class ModelMapping {
	    id: string;
	    name: string;
	    mappings: ModelMappingEntry[];
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ModelMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mappings = this.convertValues(source["mappings"], ModelMappingEntry);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Provider {
	    id: string;
	    name: string;
	    type: string;
	    authType: string;
	    apiEndpoint: string;
	    chatPath?: string;
	    headers: Record<string, string>;
	    enabled: boolean;
	    createdAt: number;
	    updatedAt: number;
	    description?: string;
	    icon?: string;
	    supportedModels?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Provider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.authType = source["authType"];
	        this.apiEndpoint = source["apiEndpoint"];
	        this.chatPath = source["chatPath"];
	        this.headers = source["headers"];
	        this.enabled = source["enabled"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.supportedModels = source["supportedModels"];
	    }
	}
	
	export class ProxyStatus {
	    isRunning: boolean;
	    port: number;
	    host: string;
	    startedAt: number;
	    uptime: number;
	    addr: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isRunning = source["isRunning"];
	        this.port = source["port"];
	        this.host = source["host"];
	        this.startedAt = source["startedAt"];
	        this.uptime = source["uptime"];
	        this.addr = source["addr"];
	    }
	}
	export class RequestLogEntry {
	    id: string;
	    timestamp: number;
	    providerId: string;
	    accountId: string;
	    model: string;
	    requestBody: Record<string, any>;
	    responseBody?: Record<string, any>;
	    statusCode: number;
	    latency: number;
	    success: boolean;
	    errorMessage?: string;
	    redacted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RequestLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.providerId = source["providerId"];
	        this.accountId = source["accountId"];
	        this.model = source["model"];
	        this.requestBody = source["requestBody"];
	        this.responseBody = source["responseBody"];
	        this.statusCode = source["statusCode"];
	        this.latency = source["latency"];
	        this.success = source["success"];
	        this.errorMessage = source["errorMessage"];
	        this.redacted = source["redacted"];
	    }
	}
	
	export class SessionRecord {
	    id: string;
	    model: string;
	    messages: ChatMessage[];
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.model = source["model"];
	        this.messages = this.convertValues(source["messages"], ChatMessage);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Statistics {
	    totalRequests: number;
	    successRequests: number;
	    failedRequests: number;
	    totalLatency: number;
	    modelUsage: Record<string, string>;
	    providerUsage: Record<string, string>;
	    accountUsage: Record<string, string>;
	    dailyStats: Record<string, string>;
	    lastUpdated: number;
	
	    static createFrom(source: any = {}) {
	        return new Statistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalRequests = source["totalRequests"];
	        this.successRequests = source["successRequests"];
	        this.failedRequests = source["failedRequests"];
	        this.totalLatency = source["totalLatency"];
	        this.modelUsage = source["modelUsage"];
	        this.providerUsage = source["providerUsage"];
	        this.accountUsage = source["accountUsage"];
	        this.dailyStats = source["dailyStats"];
	        this.lastUpdated = source["lastUpdated"];
	    }
	}
	
	

}

