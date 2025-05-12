using Repo.Functions.Examples;
using System.Net;
using static Repo.Functions.Helpers;

namespace Repo.Functions
{
    public class HttpTriggers(ILogger<HttpTriggers> logger)
    {
        private readonly ILogger logger = logger;

        // GET /api/v2/recall/sandbox/{moduleId}/info
        [Function("GetInitialInfoAsync")]
        [OpenApiOperation(tags: ["Repo"], Summary = "Get initial info")]
        [OpenApiParameter("moduleId", Required = true, Type = typeof(string), In = ParameterLocation.Path)]
        [OpenApiRequestBody("application/json", typeof(CreateRequestBody), Example = typeof(CreateRequestBodyExample))]
        [OpenApiResponseWithBody(statusCode: HttpStatusCode.OK, "application/json", typeof(ExampleResponse), Example = typeof(ExampleResponse))]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.BadRequest)]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.Unauthorized, Description = "Authorization required")]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.MethodNotAllowed)]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.InternalServerError, Description = $"OperationFailure: {nameof(GetInitialInfoAsync)}")]
        public async Task<HttpResponseData> GetInitialInfoAsync([HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "sandbox/{moduleId}/info")] HttpRequestData req, string moduleId)
        {
            logger.LogInformation("GetInitialInfoAsync Called");
            if (string.IsNullOrWhiteSpace(moduleId))
            {
                logger.LogWarning("InvalidModule - moduleId: {moduleId}", moduleId);
                return req.BadRequest("InvalidModule", "invalid module");
            }
            var body = await req.ReadBodyAs<JObject>();
            var principal = req.GetClaimsPrincipal();
            var requestInfo = CreateRequestInfo(body, principal, req, clientIpProvider, false);
            var result = await sandboxManager.GetInfo(moduleId, requestInfo);
            return req.Ok(result);
        }

        [Function("GetAsync")]
        [RequireDocsToken(OperationType.Read)]
        [OpenApiOperation(tags: ["Repo"], Summary = "Get sandbox")]
        [OpenApiSecurity("DocsToken", SecuritySchemeType.Http, BearerFormat = "Docs Token", Scheme = OpenApiSecuritySchemeType.Bearer)]
        [OpenApiParameter("moduleId", Required = true, Type = typeof(string), In = ParameterLocation.Path)]
        [OpenApiRequestBody("application/json", typeof(ResourceResult), Example = typeof(GetSandboxRequestExample))]
        [OpenApiResponseWithBody(statusCode: HttpStatusCode.OK, "application/json", typeof(ResourceResult), Example = typeof(GetSandboxResponseExample))]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.BadRequest)]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.Unauthorized, Description = "Authorization required")]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.InternalServerError, Description = $"OperationFailure: {nameof(GetAsync)}")]
        public async Task<HttpResponseData> GetAsync([HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "sandbox/{moduleId}")] HttpRequestData req, string moduleId)
        {
            logger.LogInformation("Get Resources Called");
            if (string.IsNullOrWhiteSpace(moduleId))
            {
                logger.LogWarning("InvalidModule - moduleId: {moduleId}", moduleId);
                return req.BadRequest("InvalidModule", "invalid module");
            }

            var principal = req.GetClaimsPrincipal();
            var userId = principal.GetDocsId();
            if (string.IsNullOrWhiteSpace(userId))
            {
                logger.LogWarning("MissingUserId - No userId found in cookie or token");
                return req.BadRequest("MissingUserId", "UserId is required for the user");
            }

            var userEmail = GetUserEmailFromClaims(principal, logger);
            if (string.IsNullOrWhiteSpace(userEmail))
            {
                logger.LogWarning("MissingEmail - No email found in cookie or token. userId: {userId}", userId);
                return req.BadRequest("MissingEmail", "Email is required for the user");
            }
            var body = await req.ReadBodyAs<JObject>();

            var requestInfo = CreateRequestInfo(body, principal, req, false);
            var result = await sandboxManager.CheckResourcesAsync(requestInfo);
            return req.Ok(result);
        }

        private async Task<HttpResponseData> PostAsync(HttpRequestData req, string moduleId, ClaimsPrincipal principal)
        {
            var userId = principal.GetDocsId();
            if (string.IsNullOrWhiteSpace(userId))
            {
                logger.LogWarning("MissingUserId - No userId found in cookie or token");
                return req.BadRequest("MissingUserId", "UserId is required for the user");
            }

            var userEmail = GetUserEmailFromClaims(principal, logger);
            if (string.IsNullOrWhiteSpace(userEmail))
            {
                logger.LogWarning("MissingEmail - No email found in cookie or token. userId: {userId}", userId);
                return req.BadRequest("MissingEmail", "Email is required for the user");
            }

            var body = await req.ReadBodyAs<JObject>();
            var requestInfo = CreateRequestInfo(body, principal, req, clientIpProvider, inviteOnly);
            var result = await sandboxManager.CreateResourcesAsync(userId, moduleId, userEmail, requestInfo);
            return req.Ok(result);
        }

        // POST /api/v2/recall/sandbox/preprovision/{moduleId}
        [Function("PreprovisionSandboxAsync")]
        [RequireDocsToken]
        [OpenApiOperation(tags: ["Repo"], Summary = "Preprovision sandbox")]
        [OpenApiSecurity("DocsToken", SecuritySchemeType.Http, BearerFormat = "Docs Token", Scheme = OpenApiSecuritySchemeType.Bearer)]
        [OpenApiParameter("moduleId", Required = true, Type = typeof(string), In = ParameterLocation.Path)]
        [OpenApiParameter("X-Akamai-Edgescape", Required = false, Type = typeof(string), In = ParameterLocation.Header)]
        [OpenApiParameter("X-SID", Required = false, Type = typeof(string), In = ParameterLocation.Header)]
        [OpenApiResponseWithBody(statusCode: HttpStatusCode.OK, "application/json", typeof(ResourceResult), Example = typeof(PreprovisionSandboxResponseExample))]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.BadRequest)]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.Unauthorized, Description = "Authorization required")]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.InternalServerError, Description = $"OperationFailure: {nameof(PreprovisionSandboxAsync)}")]
        public async Task<HttpResponseData> PreprovisionSandboxAsync([HttpTrigger(AuthorizationLevel.Anonymous, "post", Route = "sandbox/preprovision/{moduleId}")] HttpRequestData req, string moduleId)
        {

            logger.LogInformation("Preprovision Sandbox Called");

            if (string.IsNullOrWhiteSpace(moduleId))
            {
                logger.LogWarning("InvalidModule - moduleId: {moduleId}", moduleId);
                return req.BadRequest("InvalidModule", "invalid module");
            }
            var principal = req.GetClaimsPrincipal();
            var userId = principal.GetDocsId();
            if (string.IsNullOrWhiteSpace(userId))
            {
                logger.LogWarning("MissingUserId - No userId found in cookie or token");
                return req.BadRequest("MissingUserId", "UserId is required for the user");
            }
            var userEmail = GetUserEmailFromClaims(principal, logger);
            if (string.IsNullOrWhiteSpace(userEmail))
            {
                logger.LogWarning("MissingEmail - No email found in cookie or token. userId: {userId}", userId);
                return req.BadRequest("MissingEmail", "Email is required for the user");
            }
            var result = await PostAsync(req, moduleId, true, principal);
            auditLogger.LogUserAction("PreprovisionSandbox", OpenTelemetry.Audit.Geneva.OperationType.Create, userId, "User", userId);
            return result;
        }

        [Function("VerifyModules")]
        [OpenApiOperation(tags: ["Repo"], Summary = "Verify modules")]
        [OpenApiResponseWithBody(statusCode: HttpStatusCode.OK, "application/json", typeof(string[]), Example = typeof(VerifyModulesResponseExample))]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.InternalServerError, Description = $"OperationFailure: {nameof(VerifyModules)}")]
        public async Task<HttpResponseData> VerifyModules([HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "sandbox/verify")] HttpRequestData req)
        {
            logger.LogInformation("Verify Modules called");
            var modules = await sandboxManager.GetModulesAsync();
            return req.Ok(modules.Select(m => m.ModuleId).ToArray());
        }

        private static string GetOid(ClaimsPrincipal principal, ILogger logger)
        {
            var oid = principal.Claims.GetValueByName("oid");
            if (string.IsNullOrWhiteSpace(oid))
            {
                logger.LogWarning("MissingOid - No oid found in cookie or token");
                throw new DocsApiException(HttpStatusCode.BadRequest, "MissingOid", "Oid is required for the user");
            }
            return oid;
        }
        private static string GetEmail(ClaimsPrincipal principal, ILogger logger, string oid)
        {
            var userEmail = GetUserEmailFromClaims(principal, logger);
            if (string.IsNullOrWhiteSpace(userEmail))
            {
                logger.LogWarning("MissingEmail - No email found in cookie or token. Oid: {oid}", oid);
                throw new DocsApiException(HttpStatusCode.BadRequest, "MissingEmail", "Email is required for the user");
            }
            return userEmail;
        }
    }
}
