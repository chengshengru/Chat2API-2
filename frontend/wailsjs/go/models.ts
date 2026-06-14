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
	export class PaginatedResult {
	    items: LogEntry[];
	    total: number;
	    page: number;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new PaginatedResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], LogEntry);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.size = source["size"];
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
	export class PersistentStatistics {
	    totalRequests: number;
	    successRequests: number;
	    failedRequests: number;
	    totalLatency: number;
	    lastUpdated: number;
	    modelUsage: Record<string, string>;
	    providerUsage: Record<string, string>;
	    accountUsage: Record<string, string>;
	    dailyStats: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new PersistentStatistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalRequests = source["totalRequests"];
	        this.successRequests = source["successRequests"];
	        this.failedRequests = source["failedRequests"];
	        this.totalLatency = source["totalLatency"];
	        this.lastUpdated = source["lastUpdated"];
	        this.modelUsage = source["modelUsage"];
	        this.providerUsage = source["providerUsage"];
	        this.accountUsage = source["accountUsage"];
	        this.dailyStats = source["dailyStats"];
	    }
	}

}

export namespace oauth {
	
	export class AccountInfo {
	    userId?: string;
	    name?: string;
	    email?: string;
	    quota?: number;
	    used?: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userId = source["userId"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.quota = source["quota"];
	        this.used = source["used"];
	    }
	}
	export class OAuthResult {
	    success: boolean;
	    providerId?: string;
	    providerType?: string;
	    credentials?: Record<string, string>;
	    account?: types.Account;
	    accountInfo?: AccountInfo;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new OAuthResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.providerId = source["providerId"];
	        this.providerType = source["providerType"];
	        this.credentials = source["credentials"];
	        this.account = this.convertValues(source["account"], types.Account);
	        this.accountInfo = this.convertValues(source["accountInfo"], AccountInfo);
	        this.error = source["error"];
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
	export class TokenValidationResult {
	    valid: boolean;
	    tokenType?: string;
	    expiresAt?: number;
	    accountInfo?: AccountInfo;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TokenValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.tokenType = source["tokenType"];
	        this.expiresAt = source["expiresAt"];
	        this.accountInfo = this.convertValues(source["accountInfo"], AccountInfo);
	        this.error = source["error"];
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

}

export namespace proxy {
	
	export class ProxyStatus {
	    isRunning: boolean;
	    port: number;
	    host: string;
	    uptime: number;
	    startedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ProxyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isRunning = source["isRunning"];
	        this.port = source["port"];
	        this.host = source["host"];
	        this.uptime = source["uptime"];
	        this.startedAt = source["startedAt"];
	    }
	}
	export class Statistics {
	    totalRequests: number;
	    successRequests: number;
	    failedRequests: number;
	    activeConnections: number;
	    totalLatency: number;
	    lastUpdated: number;
	    modelUsage: Record<string, number>;
	    providerUsage: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new Statistics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalRequests = source["totalRequests"];
	        this.successRequests = source["successRequests"];
	        this.failedRequests = source["failedRequests"];
	        this.activeConnections = source["activeConnections"];
	        this.totalLatency = source["totalLatency"];
	        this.lastUpdated = source["lastUpdated"];
	        this.modelUsage = source["modelUsage"];
	        this.providerUsage = source["providerUsage"];
	    }
	}

}

export namespace types {
	
	export class Account {
	    id: string;
	    providerId: string;
	    name: string;
	    email?: string;
	    credentials?: Record<string, string>;
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
	    name: string;
	    key: string;
	    enabled: boolean;
	    createdAt: number;
	    lastUsedAt?: number;
	    usageCount: number;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new ApiKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.key = source["key"];
	        this.enabled = source["enabled"];
	        this.createdAt = source["createdAt"];
	        this.lastUsedAt = source["lastUsedAt"];
	        this.usageCount = source["usageCount"];
	        this.description = source["description"];
	    }
	}
	export class ManagementApiConfig {
	    enableManagementApi: boolean;
	    managementApiSecret: string;
	    managementApiPort?: number;
	
	    static createFrom(source: any = {}) {
	        return new ManagementApiConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enableManagementApi = source["enableManagementApi"];
	        this.managementApiSecret = source["managementApiSecret"];
	        this.managementApiPort = source["managementApiPort"];
	    }
	}
	export class ToolCallingConfig {
	    enabled: boolean;
	    promptMode: string;
	    customTemplate?: string;
	    maxToolsPerRequest: number;
	    maxRetries: number;
	    timeoutSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new ToolCallingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.promptMode = source["promptMode"];
	        this.customTemplate = source["customTemplate"];
	        this.maxToolsPerRequest = source["maxToolsPerRequest"];
	        this.maxRetries = source["maxRetries"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	    }
	}
	export class SessionConfig {
	    sessionTimeout: number;
	    maxMessagesPerSession: number;
	    deleteAfterTimeout: boolean;
	    maxSessionsPerAccount: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionTimeout = source["sessionTimeout"];
	        this.maxMessagesPerSession = source["maxMessagesPerSession"];
	        this.deleteAfterTimeout = source["deleteAfterTimeout"];
	        this.maxSessionsPerAccount = source["maxSessionsPerAccount"];
	    }
	}
	export class RequestLogConfig {
	    enabled: boolean;
	    maxEntries: number;
	    includeBodies: boolean;
	    maxBodyChars: number;
	    redactSensitiveData: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RequestLogConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.maxEntries = source["maxEntries"];
	        this.includeBodies = source["includeBodies"];
	        this.maxBodyChars = source["maxBodyChars"];
	        this.redactSensitiveData = source["redactSensitiveData"];
	    }
	}
	export class ModelMapping {
	    requestModel: string;
	    actualModel: string;
	    preferredProviderId?: string;
	    preferredAccountId?: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestModel = source["requestModel"];
	        this.actualModel = source["actualModel"];
	        this.preferredProviderId = source["preferredProviderId"];
	        this.preferredAccountId = source["preferredAccountId"];
	    }
	}
	export class ModelMappingEntry {
	    key: string;
	    value: ModelMapping;
	
	    static createFrom(source: any = {}) {
	        return new ModelMappingEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = this.convertValues(source["value"], ModelMapping);
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
	    proxyPort: number;
	    proxyHost: string;
	    loadBalanceStrategy: string;
	    modelMappings: ModelMappingEntry[];
	    theme: string;
	    autoStart: boolean;
	    autoStartProxy: boolean;
	    minimizeToTray: boolean;
	    logLevel: string;
	    logRetentionDays: number;
	    requestLogConfig: RequestLogConfig;
	    requestTimeout: number;
	    retryCount: number;
	    apiKeys: ApiKey[];
	    enableApiKey: boolean;
	    oauthProxyMode: string;
	    sessionConfig: SessionConfig;
	    toolCallingConfig: ToolCallingConfig;
	    managementApi: ManagementApiConfig;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyPort = source["proxyPort"];
	        this.proxyHost = source["proxyHost"];
	        this.loadBalanceStrategy = source["loadBalanceStrategy"];
	        this.modelMappings = this.convertValues(source["modelMappings"], ModelMappingEntry);
	        this.theme = source["theme"];
	        this.autoStart = source["autoStart"];
	        this.autoStartProxy = source["autoStartProxy"];
	        this.minimizeToTray = source["minimizeToTray"];
	        this.logLevel = source["logLevel"];
	        this.logRetentionDays = source["logRetentionDays"];
	        this.requestLogConfig = this.convertValues(source["requestLogConfig"], RequestLogConfig);
	        this.requestTimeout = source["requestTimeout"];
	        this.retryCount = source["retryCount"];
	        this.apiKeys = this.convertValues(source["apiKeys"], ApiKey);
	        this.enableApiKey = source["enableApiKey"];
	        this.oauthProxyMode = source["oauthProxyMode"];
	        this.sessionConfig = this.convertValues(source["sessionConfig"], SessionConfig);
	        this.toolCallingConfig = this.convertValues(source["toolCallingConfig"], ToolCallingConfig);
	        this.managementApi = this.convertValues(source["managementApi"], ManagementApiConfig);
	        this.language = source["language"];
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
	    modelMappings?: Record<string, string>;
	    status?: string;
	    lastStatusCheck?: number;
	
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
	        this.modelMappings = source["modelMappings"];
	        this.status = source["status"];
	        this.lastStatusCheck = source["lastStatusCheck"];
	    }
	}
	
	
	export class SessionRecord {
	    id: string;
	    providerId: string;
	    accountId: string;
	    providerType?: string;
	    sessionKey?: string;
	    providerSessionId: string;
	    parentMessageId?: string;
	    sessionType: string;
	    messages: string[];
	    credentials?: Record<string, string>;
	    createdAt: number;
	    lastActiveAt: number;
	    status: string;
	    model?: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.providerId = source["providerId"];
	        this.accountId = source["accountId"];
	        this.providerType = source["providerType"];
	        this.sessionKey = source["sessionKey"];
	        this.providerSessionId = source["providerSessionId"];
	        this.parentMessageId = source["parentMessageId"];
	        this.sessionType = source["sessionType"];
	        this.messages = source["messages"];
	        this.credentials = source["credentials"];
	        this.createdAt = source["createdAt"];
	        this.lastActiveAt = source["lastActiveAt"];
	        this.status = source["status"];
	        this.model = source["model"];
	    }
	}

}

