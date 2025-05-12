using Repo.API.Examples;
using Repo.API.Helpers;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using System.Net;
using System.Threading.Tasks;

namespace Repo.API
{
    public class DashboardApi(ITelemetryEventTracker telemetry, IDashboardService dashboardService, IAuditLogger auditLogger, ILogger<DashboardApi> logger)
    {
        private readonly IAuditLogger auditLogger = auditLogger;
        private readonly ILogger logger = logger;

        [Function(nameof(GetDashboardSummary))]
        [RequireDocsToken(OperationType.Read)]
        [OpenApiOperation(tags: ["Dashboard"], Summary = "Get dashboard summary")]
        [OpenApiSecurity("DocsToken", SecuritySchemeType.Http, BearerFormat = "Docs Token", Scheme = OpenApiSecuritySchemeType.Bearer)]
        [OpenApiParameter("locale", Required = false, Type = typeof(string), In = ParameterLocation.Query)]
        [OpenApiParameter("itemLimit", Required = false, Type = typeof(int), In = ParameterLocation.Query)]
        [OpenApiResponseWithBody(statusCode: HttpStatusCode.OK, "application/json", typeof(DashboardSummary), Example = typeof(DashboardSummaryExample))]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.Unauthorized, Description = "Authorization required")]
        [OpenApiResponseWithoutBody(statusCode: HttpStatusCode.InternalServerError, Description = $"OperationFailure: {nameof(GetDashboardSummary)}")]
        public async Task<HttpResponseData> GetDashboardSummary(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "dashboard-summary/{param}")]
            HttpRequestData req,
            FunctionContext context)
        {

            logger.LogInformation($"{context.FunctionDefinition.Name} called");

            var locale = req.Query.Get("locale")?.ToString().ToLower();
            var normalizedLocale = LocaleHelper.Normalize(locale);

            var result = await dashboardService.GetDashboardSummary();

            if (!result.IsCertificationLinked) telemetry.LinkedAccountNotFound(docsId);

            return req.Ok(result);
        }
    }
}
