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

        [Function(nameof(GetDashboardSummary))]
        [RequireDocsToken(OperationType.Read)]
        public async Task<HttpResponseData> GetDashboardSummary([HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "dashboard-summary")] HttpRequestData req, FunctionContext context)
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
