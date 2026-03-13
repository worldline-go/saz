<script lang="ts">
  import { onMount, untrack } from "svelte";
  import {
    requestProcesses,
    requestProcessTerminate,
    type ProcessQueryParams,
  } from "@/helper/call";
  import { addToast } from "@/store/toast";
  import {
    RefreshCw,
    XCircle,
    ChevronRight,
    ChevronDown,
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
    ChevronLeft,
    Search,
  } from "@lucide/svelte";
  import type { process, processStatus } from "@/helper/model";

  let processes = $state<process[]>([]);
  let loading = $state(false);
  let expandedIds = $state<Set<string>>(new Set());

  // Filter / sort / pagination state
  let statusFilter = $state<string>("");
  let sortField = $state<string>("created_at");
  let sortDir = $state<"asc" | "desc">("desc");
  let limit = $state(25);
  let offset = $state(0);

  // Dedicated filter inputs
  let userFilter = $state("");
  let dateFrom = $state("");
  let dateTo = $state("");
  let dbFilter = $state("");

  // Raw server-side query (appended to the API request as-is)
  let rawQuery = $state("");

  function buildParams(): ProcessQueryParams {
    const params: ProcessQueryParams = {
      sort: sortDir === "desc" ? `-${sortField}` : sortField,
      limit,
      offset,
    };
    if (statusFilter) params.status = statusFilter;
    if (userFilter.trim()) params.user = userFilter.trim();
    if (dateFrom.trim()) params.dateFrom = dateFrom.trim();
    if (dateTo.trim()) params.dateTo = dateTo.trim();
    if (dbFilter.trim()) params.database = dbFilter.trim();
    if (rawQuery.trim()) params.rawQuery = rawQuery.trim();
    return params;
  }

  const fetchProcesses = async () => {
    loading = true;
    processes = await requestProcesses(buildParams());
    loading = false;
  };

  const terminateProcess = async (pid: string) => {
    try {
      await requestProcessTerminate(pid);
      addToast("Process terminated", "info");
      await fetchProcesses();
    } catch (error: any) {
      const msg =
        error?.response?.data?.error || error?.message || "Unknown error";
      addToast("Error terminating process: " + msg, "alert");
    }
  };

  const toggleExpand = (id: string) => {
    const next = new Set(expandedIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    expandedIds = next;
  };

  const toggleSort = (field: string) => {
    if (sortField === field) {
      sortDir = sortDir === "asc" ? "desc" : "asc";
    } else {
      sortField = field;
      sortDir = "desc";
    }
    offset = 0;
    fetchProcesses();
  };

  const prevPage = () => {
    offset = Math.max(0, offset - limit);
    fetchProcesses();
  };
  const nextPage = () => {
    if (processes.length === limit) {
      offset += limit;
      fetchProcesses();
    }
  };

  const statusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-blue-100 text-blue-700";
      case "completed":
        return "bg-green-100 text-green-700";
      case "failed":
        return "bg-red-100 text-red-700";
      case "terminated":
        return "bg-gray-200 text-gray-600";
      default:
        return "bg-gray-100 text-gray-500";
    }
  };

  const formatDate = (date: string) => {
    if (!date) return "-";
    return new Date(date).toLocaleString();
  };

  const sortIcon = (field: string): "neutral" | "asc" | "desc" => {
    if (sortField !== field) return "neutral";
    return sortDir;
  };

  const statuses: { label: string; value: string }[] = [
    { label: "All", value: "" },
    { label: "Running", value: "running" },
    { label: "Completed", value: "completed" },
    { label: "Failed", value: "failed" },
    { label: "Terminated", value: "terminated" },
  ];

  let initialized = $state(false);

  $effect(() => {
    // re-fetch when status filter changes (skip initial mount)
    statusFilter;
    if (initialized) {
      untrack(() => {
        offset = 0;
        fetchProcesses();
      });
    }
  });

  const applyFilters = () => { offset = 0; fetchProcesses(); };
  const filterKeydown = (e: KeyboardEvent) => { if (e.key === "Enter") applyFilters(); };

  onMount(() => {
    fetchProcesses().then(() => {
      initialized = true;
    });
  });
</script>

<div class="flex flex-col h-full w-full overflow-hidden text-sm">
  <!-- Toolbar -->
  <div class="border-b border-gray-300 flex items-center justify-between gap-2 px-3 h-8 shrink-0 bg-gray-100">
    <div class="flex items-center gap-2">
      <span class="font-medium text-gray-700">Processes</span>
      <span class="border-l border-gray-300 h-4"></span>
      <!-- Status filter -->
      <select
        class="border-none bg-transparent text-sm hover:cursor-pointer hover:bg-white h-8 px-1"
        bind:value={statusFilter}
      >
        {#each statuses as s}
          <option value={s.value}>{s.label}</option>
        {/each}
      </select>
    </div>
    <div class="flex items-center">
      <button
        class="px-2 h-8 hover:cursor-pointer hover:bg-gray-200 text-gray-600"
        onclick={fetchProcesses}
        disabled={loading}
        title="Refresh"
      >
        <RefreshCw class={loading ? "animate-spin" : ""} size={16} />
      </button>
    </div>
  </div>
  <!-- Filter row -->
  <div class="border-b border-gray-200 flex items-center gap-3 px-3 h-9 shrink-0 bg-gray-50 text-xs text-gray-600">
    <label class="flex items-center gap-1">
      <span>User</span>
      <input
        type="text"
        class="border border-gray-300 rounded bg-white h-6 px-1.5 w-28 text-xs"
        placeholder="admin"
        bind:value={userFilter}
        onkeydown={filterKeydown}
      />
    </label>
    <label class="flex items-center gap-1">
      <span>From</span>
      <input
        type="date"
        class="border border-gray-300 rounded bg-white h-6 px-1.5 text-xs"
        bind:value={dateFrom}
        onkeydown={filterKeydown}
      />
    </label>
    <label class="flex items-center gap-1">
      <span>To</span>
      <input
        type="date"
        class="border border-gray-300 rounded bg-white h-6 px-1.5 text-xs"
        bind:value={dateTo}
        onkeydown={filterKeydown}
      />
    </label>
    <label class="flex items-center gap-1">
      <span>DB</span>
      <input
        type="text"
        class="border border-gray-300 rounded bg-white h-6 px-1.5 w-28 text-xs"
        placeholder="postgres"
        bind:value={dbFilter}
        onkeydown={filterKeydown}
      />
    </label>
    <span class="border-l border-gray-300 h-4"></span>
    <input
      type="text"
      class="border border-gray-300 rounded bg-white h-6 px-1.5 flex-1 min-w-0 text-xs font-mono"
      placeholder="raw query"
      title="Raw query appended to request (Enter to apply)"
      bind:value={rawQuery}
      onkeydown={filterKeydown}
    />
    <button
      class="px-1.5 h-6 rounded hover:cursor-pointer hover:bg-gray-200 text-gray-500"
      onclick={applyFilters}
      disabled={loading}
      title="Apply filters"
    >
      <Search size={12} />
    </button>
  </div>

  <!-- Table -->
  <div class="flex-1 overflow-y-auto min-h-0">
    <table class="w-full text-sm">
      <thead class="sticky top-0 bg-gray-50 z-10">
        <tr class="border-b border-gray-300 text-left text-gray-500">
          <th class="px-3 py-2 w-8"></th>
          <th class="px-3 py-2 w-24">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("status")}>
              Status
              {#if sortIcon("status") === "asc"}<ArrowUp size={12} />{:else if sortIcon("status") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2">Note</th>
          <th class="px-3 py-2">Description</th>
          <th class="px-3 py-2 w-28">Database</th>
          <th class="px-3 py-2 w-24">Driver</th>
          <th class="px-3 py-2 w-28">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("info.duration")}>
              Duration
              {#if sortIcon("info.duration") === "asc"}<ArrowUp size={12} />{:else if sortIcon("info.duration") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2 w-20">Rows</th>
          <th class="px-3 py-2 w-24">User</th>
          <th class="px-3 py-2 w-44">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("created_at")}>
              Created
              {#if sortIcon("created_at") === "asc"}<ArrowUp size={12} />{:else if sortIcon("created_at") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2 w-44">
            <button class="flex items-center gap-1 hover:text-gray-800 hover:cursor-pointer" onclick={() => toggleSort("updated_at")}>
              Updated
              {#if sortIcon("updated_at") === "asc"}<ArrowUp size={12} />{:else if sortIcon("updated_at") === "desc"}<ArrowDown size={12} />{:else}<ArrowUpDown size={12} />{/if}
            </button>
          </th>
          <th class="px-3 py-2 w-20">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#if processes.length === 0 && !loading}
          <tr>
            <td colspan="12" class="text-gray-400 text-center py-10">No processes found</td>
          </tr>
        {/if}
        {#each processes as proc (proc.id)}
          {@const isExpanded = expandedIds.has(proc.id)}
          {@const hasDetail = proc.info.query || proc.info.error || (proc.info.cells && proc.info.cells.length > 0)}
          <tr
            class={[
              "border-b border-gray-200 transition-colors duration-100",
              hasDetail ? "hover:bg-gray-50 cursor-pointer" : "",
              isExpanded ? "bg-gray-50" : "",
            ]}
            onclick={() => { if (hasDetail) toggleExpand(proc.id); }}
          >
            <td class="px-3 py-2.5 text-gray-400">
              {#if hasDetail}
                {#if isExpanded}
                  <ChevronDown size={14} />
                {:else}
                  <ChevronRight size={14} />
                {/if}
              {/if}
            </td>
            <td class="px-3 py-2.5">
              <span class={["px-2 py-0.5 text-xs font-medium rounded", statusColor(proc.status)]}>
                {proc.status}
              </span>
            </td>
            <td class="px-3 py-2.5 max-w-56">
              <div class="cell-truncate" title={proc.info.note}>
                {proc.info.note || "-"}
              </div>
            </td>
            <td class="px-3 py-2.5 max-w-64">
              <div class="cell-truncate" title={proc.info.description}>
                {proc.info.description || "-"}
              </div>
            </td>
            <td class="px-3 py-2.5 max-w-28">
              <div class="cell-truncate" title={proc.info.database}>
                {proc.info.database || "-"}
              </div>
            </td>
            <td class="px-3 py-2.5 max-w-24">
              <div class="cell-truncate" title={proc.info.driver}>
                {proc.info.driver || "-"}
              </div>
            </td>
            <td class="px-3 py-2.5 tabular-nums">
              {proc.info.duration || "-"}
            </td>
            <td class="px-3 py-2.5 tabular-nums">
              {proc.info.rows_affected ?? "-"}
            </td>
            <td class="px-3 py-2.5 max-w-24">
              <div class="cell-truncate" title={proc.user}>
                {proc.user || "-"}
              </div>
            </td>
            <td class="px-3 py-2.5 whitespace-nowrap tabular-nums">
              {formatDate(proc.created_at)}
            </td>
            <td class="px-3 py-2.5 whitespace-nowrap tabular-nums">
              {formatDate(proc.updated_at)}
            </td>
            <td class="px-3 py-2.5">
              {#if proc.status === "running"}
                <button
                  class="text-red-500 hover:text-white hover:bg-red-500 p-1 rounded hover:cursor-pointer"
                  onclick={(e: MouseEvent) => { e.stopPropagation(); terminateProcess(proc.id); }}
                  title="Terminate"
                >
                  <XCircle size={16} />
                </button>
              {/if}
            </td>
          </tr>
          {#if isExpanded}
            <tr class="border-b border-gray-200">
              <td colspan="12" class="p-0">
                <div class="border-l-2 border-gray-300 ml-5 my-2">
                  {#if proc.info.error}
                    <div class="px-4 py-2 bg-red-50 text-red-700 text-sm">
                      <span class="font-medium">Error:</span> {proc.info.error}
                    </div>
                  {/if}
                  {#if proc.info.query}
                    <div class="px-4 py-2">
                      <span class="text-gray-400 text-xs uppercase tracking-wide">Query</span>
                      <pre class="mt-1 text-sm text-gray-700 whitespace-pre-wrap break-all bg-white border border-gray-200 rounded px-3 py-2 max-h-72 overflow-y-auto">{proc.info.query}</pre>
                    </div>
                  {/if}
                  {#if proc.info.cells && proc.info.cells.length > 0}
                    <div class="px-4 py-2">
                      <span class="text-gray-400 text-xs uppercase tracking-wide">Cells</span>
                      <table class="mt-1 w-full text-sm border border-gray-200 rounded overflow-hidden">
                        <thead>
                          <tr class="bg-gray-50 text-left text-gray-500 text-xs">
                            <th class="px-3 py-1.5 w-8">#</th>
                            <th class="px-3 py-1.5">Description</th>
                            <th class="px-3 py-1.5 w-28">Database</th>
                            <th class="px-3 py-1.5 w-24">Driver</th>
                            <th class="px-3 py-1.5 w-24">Status</th>
                            <th class="px-3 py-1.5 w-28">Duration</th>
                            <th class="px-3 py-1.5 w-20">Rows</th>
                          </tr>
                        </thead>
                        <tbody>
                          {#each proc.info.cells as cell, idx}
                            {@const cellStatusColor =
                              cell.status === "completed" ? "bg-green-100 text-green-700" :
                              cell.status === "failed" ? "bg-red-100 text-red-700" :
                              cell.status === "skipped" ? "bg-gray-100 text-gray-500" :
                              "bg-gray-100 text-gray-500"
                            }
                            <tr class="border-t border-gray-200">
                              <td class="px-3 py-1.5 text-gray-400 tabular-nums">{idx + 1}</td>
                              <td class="px-3 py-1.5">
                                <div class="cell-truncate max-w-md" title={cell.description}>{cell.description || "-"}</div>
                                {#if cell.query}
                                  <pre class="mt-1 text-xs text-gray-500 whitespace-pre-wrap break-all max-h-24 overflow-y-auto">{cell.query}</pre>
                                {/if}
                                {#if cell.error}
                                  <div class="mt-1 text-xs text-red-600">{cell.error}</div>
                                {/if}
                              </td>
                              <td class="px-3 py-1.5">{cell.database || "-"}</td>
                              <td class="px-3 py-1.5">{cell.driver || "-"}</td>
                              <td class="px-3 py-1.5">
                                <span class={["px-2 py-0.5 text-xs font-medium rounded", cellStatusColor]}>
                                  {cell.status}
                                </span>
                              </td>
                              <td class="px-3 py-1.5 tabular-nums">{cell.duration || "-"}</td>
                              <td class="px-3 py-1.5 tabular-nums">{cell.rows_affected ?? "-"}</td>
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  {/if}
                </div>
              </td>
            </tr>
          {/if}
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  <div class="border-t border-gray-300 flex items-center justify-between px-3 h-10 shrink-0 bg-gray-100 text-sm text-gray-500">
    <div class="flex items-center gap-2">
      <span>Showing {offset + 1}-{offset + processes.length}</span>
      <span class="border-l border-gray-300 h-4"></span>
      <label class="flex items-center gap-1">
        Per page
        <select
          class="border-none bg-transparent hover:cursor-pointer hover:bg-white h-7 text-sm px-1"
          bind:value={limit}
          onchange={() => { offset = 0; fetchProcesses(); }}
        >
          <option value={10}>10</option>
          <option value={25}>25</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </select>
      </label>
    </div>
    <div class="flex items-center gap-1">
      <button
        class="px-2 h-7 hover:bg-gray-200 hover:cursor-pointer disabled:opacity-30 disabled:cursor-default rounded"
        disabled={offset === 0}
        onclick={prevPage}
        title="Previous page"
      >
        <ChevronLeft size={14} />
      </button>
      <button
        class="px-2 h-7 hover:bg-gray-200 hover:cursor-pointer disabled:opacity-30 disabled:cursor-default rounded"
        disabled={processes.length < limit}
        onclick={nextPage}
        title="Next page"
      >
        <ChevronRight size={14} />
      </button>
    </div>
  </div>
</div>

<style>
  .cell-truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .cell-truncate:hover {
    overflow: visible;
    white-space: normal;
    word-wrap: break-word;
    position: relative;
    z-index: 20;
    background: #fffacd;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    border-radius: 4px;
    padding: 4px 6px;
    margin: -4px -6px;
  }
</style>
