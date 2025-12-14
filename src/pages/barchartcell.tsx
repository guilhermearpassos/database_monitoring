import {DataFrame, Field, FieldType} from '@grafana/data';


export interface WaitTypeData {
    type: string;
    count: number;
}

const groupToBaseColor: Record<string, string> = {
    "Locks": "239, 68, 68, 1",   // Red
    "I/O": "249, 115, 22, 1",    // Orange
    "CPU": "34, 197, 94, 1",     // Green
    "Memory": "234, 179, 8, 1",  // Yellow
    "Network": "168, 85, 247, 1", // Purple
    "Background": "59, 130, 246, 1", // Blue
    "Idle": "107, 114, 128, 1",  // Gray
    "Miscellaneous": "132, 204, 22, 1", // Lime
};

const waitEventToIntensity: Record<string, number> = {
    // Locks and Blocking
    "LCK_M_BU": 0.2,
    "LCK_M_IS": 0.25,
    "LCK_M_IU": 0.3,
    "LCK_M_S": 0.35,
    "LCK_M_IX": 0.4,
    "LCK_M_X": 0.45,
    "LCK_M_U": 0.5,
    "LCK_M_SIU": 0.55,
    "LCK_M_SIX": 0.6,
    "LCK_M_UIX": 0.65,

    // I/O-Related
    "ASYNC_IO_COMPLETION": 0.2,
    "IO_COMPLETION": 0.25,
    "PAGEIOLATCH_SH": 0.3,
    "PAGEIOLATCH_EX": 0.35,
    "BACKUPIO": 0.4,
    "WRITELOG": 0.45,
    "LOGBUFFER": 0.5,

    // CPU and Parallelism
    "CXPACKET": 0.2,
    "SOS_SCHEDULER_YIELD": 0.25,
    "THREADPOOL": 0.3,
    "RESOURCE_SEMAPHORE": 0.35,
    "RESOURCE_SEMAPHORE_QUERY_COMPILE": 0.4,
    "cpu": 0.45,

    // Memory-Related
    "CMEMTHREAD": 0.2,
    "MEMORY_ALLOCATION_EXT": 0.25,
    "PAGELATCH_EX": 0.3,
    "PAGELATCH_SH": 0.35,
    "RESERVED_MEMORY_ALLOCATION_EXT": 0.4,

    // Network and Latency
    "ASYNC_NETWORK_IO": 0.2,
    "NETWORK_IO": 0.25,
    "OLEDB": 0.3,

    // Background and Maintenance
    "CHECKPOINT_QUEUE": 0.2,
    "LAZYWRITER_SLEEP": 0.25,
    "XE_TIMER_EVENT": 0.3,
    "TRACEWRITE": 0.35,
    "FT_IFTS_SCHEDULER_IDLE_WAIT": 0.4,

    // Idle and Sleep
    "SLEEP_TASK": 0.2,
    "WAITFOR": 0.25,
    "BROKER_RECEIVE_WAITFOR": 0.3,
    "BROKER_TO_FLUSH": 0.35,
    "BROKER_TRANSMITTER": 0.4,

    // Miscellaneous
    "PREEMPTIVE_OS_AUTHENTICATIONOPS": 0.2,
    "PREEMPTIVE_OS_GETPROCADDRESS": 0.25,
    "CLR_AUTO_EVENT": 0.3,
    "CLR_CRST": 0.35,
    "CLR_JOIN": 0.4,
    "CLR_MANUAL_EVENT": 0.45,
};

const waitEventToGroup: Record<string, string> = {
    // Locks and Blocking
    "LCK_M_S": "Locks",
    "LCK_M_X": "Locks",
    "LCK_M_U": "Locks",
    "LCK_M_IS": "Locks",
    "LCK_M_IU": "Locks",
    "LCK_M_IX": "Locks",
    "LCK_M_SIU": "Locks",
    "LCK_M_SIX": "Locks",
    "LCK_M_UIX": "Locks",
    "LCK_M_BU": "Locks",

    // I/O-Related
    "PAGEIOLATCH_SH": "I/O",
    "PAGEIOLATCH_EX": "I/O",
    "WRITELOG": "I/O",
    "ASYNC_IO_COMPLETION": "I/O",
    "IO_COMPLETION": "I/O",
    "BACKUPIO": "I/O",
    "LOGBUFFER": "I/O",

    // CPU and Parallelism
    "cpu": "CPU",
    "CXPACKET": "CPU",
    "SOS_SCHEDULER_YIELD": "CPU",
    "THREADPOOL": "CPU",
    "RESOURCE_SEMAPHORE": "CPU",
    "RESOURCE_SEMAPHORE_QUERY_COMPILE": "CPU",

    // Memory-Related
    "CMEMTHREAD": "Memory",
    "MEMORY_ALLOCATION_EXT": "Memory",
    "PAGELATCH_EX": "Memory",
    "PAGELATCH_SH": "Memory",
    "RESERVED_MEMORY_ALLOCATION_EXT": "Memory",

    // Network and Latency
    "ASYNC_NETWORK_IO": "Network",
    "NETWORK_IO": "Network",
    "OLEDB": "Network",

    // Background and Maintenance
    "CHECKPOINT_QUEUE": "Background",
    "LAZYWRITER_SLEEP": "Background",
    "XE_TIMER_EVENT": "Background",
    "TRACEWRITE": "Background",
    "FT_IFTS_SCHEDULER_IDLE_WAIT": "Background",

    // Idle and Sleep
    "SLEEP_TASK": "Idle",
    "WAITFOR": "Idle",
    "BROKER_RECEIVE_WAITFOR": "Idle",
    "BROKER_TO_FLUSH": "Idle",
    "BROKER_TRANSMITTER": "Idle",

    // Miscellaneous
    "PREEMPTIVE_OS_AUTHENTICATIONOPS": "Miscellaneous",
    "PREEMPTIVE_OS_GETPROCADDRESS": "Miscellaneous",
    "CLR_AUTO_EVENT": "Miscellaneous",
    "CLR_CRST": "Miscellaneous",
    "CLR_JOIN": "Miscellaneous",
    "CLR_MANUAL_EVENT": "Miscellaneous",
};

/**
 * Get the color for a wait event based on its group and intensity
 * @param waitEvent The wait event name
 * @returns RGBA color string (e.g., "239, 68, 68, 0.40")
 */
export function getWaitEventColor(waitEvent: string): string {
    // Get the group for the wait event
    const group = waitEventToGroup[waitEvent];
    if (!group) {
        return "107, 114, 128, 1"; // Default gray color
    }

    // Get the base color for the group
    const rgba = groupToBaseColor[group];
    if (!rgba) {
        return "107, 114, 128, 1"; // Default gray color
    }

    // Get the intensity level for the wait event
    let alpha = waitEventToIntensity[waitEvent];
    if (alpha === undefined) {
        alpha = 0.5; // Default intensity
    }

    // Parse the base color
    const parts = rgba.split(",");
    if (parts.length !== 4) {
        return "107, 114, 128, 1"; // Default gray color
    }

    // Extract RGB values
    const r = parts[0].trim();
    const g = parts[1].trim();
    const b = parts[2].trim();

    // Generate the RGBA color
    return `${r}, ${g}, ${b}, ${alpha.toFixed(2)}`;
}

/**
 * Get the hex color for a wait event (for use in HTML/CSS)
 * @param waitEvent The wait event name
 * @returns Hex color string with alpha (e.g., "#ef4444cc")
 */
export function getWaitEventColorHex(waitEvent: string): string {
    const rgba = getWaitEventColor(waitEvent);
    const parts = rgba.split(",").map(p => p.trim());

    if (parts.length !== 4) {
        return "#6b7280"; // Default gray
    }

    const r = parseInt(parts[0], 10);
    const g = parseInt(parts[1], 10);
    const b = parseInt(parts[2], 10);
    const a = Math.round(parseFloat(parts[3]) * 255);

    const toHex = (n: number) => n.toString(16).padStart(2, '0');

    return `#${toHex(r)}${toHex(g)}${toHex(b)}${toHex(a)}`;
}

/**
 * Generate HTML for a stacked bar chart showing wait type distribution
 */
export function generateWaitTypeHTML(waitTypes: WaitTypeData[]): string {
    if (!waitTypes || waitTypes.length === 0) {
        return '<span style="color:#888;">No data</span>';
    }

    const total = waitTypes.reduce((sum, wt) => sum + wt.count, 0);

    if (total === 0) {
        return '<span style="color:#888;">0 connections</span>';
    }

    // Sort by count descending
    const sorted = [...waitTypes].sort((a, b) => b.count - a.count);

    // Generate bar segments
    const segments = sorted
        .filter(wt => (wt.count / total) >= 0.005) // Only show segments > 0.5%
        .map(wt => {
            const percentage = (wt.count / total) * 100;
            let waitEvent = wt.type;
            if (waitEvent ==="") {
                waitEvent = "cpu"
            }
            const color = getWaitEventColor(waitEvent);
            console.warn(color)
            return `<div style="width:${percentage}%;background-color:rgba(${color});" title="${waitEvent}: ${wt.count} (${percentage.toFixed(1)}%)"></div>`;
        })
        .join('');

    // Create the stacked bar with a label

    return `
<div style="display:flex;flex-direction:column;gap:4px;">
    <div style="display:flex;height:20px;border-radius:4px;overflow:hidden;background:#2c2c2c;">
        ${segments}
    </div>
</div>
  `;
}

/**
 * Generate a simpler inline bar (no label)
 */
export function generateSimpleWaitTypeBar(waitTypes: WaitTypeData[]): string {
    if (!waitTypes || waitTypes.length === 0) {
        return '';
    }

    const total = waitTypes.reduce((sum, wt) => sum + wt.count, 0);
    if (total === 0) {
        return '';
    }

    const sorted = [...waitTypes].sort((a, b) => b.count - a.count);

    const segments = sorted
        .filter(wt => (wt.count / total) >= 0.005)
        .map(wt => {
            const percentage = (wt.count / total) * 100;
            const color = getWaitEventColorHex(wt.type);
            return `<div style="width:${percentage}%;background:${color};" title="${wt.type}: ${wt.count}"></div>`;
        })
        .join('');

    return `<div style="display:flex;height:20px;border-radius:4px;overflow:hidden;background:#2c2c2c;">${segments}</div>`;
}

/**
 * Transform a DataFrame to include HTML-rendered wait type bars
 */
export function addWaitTypeHTMLColumn(
    frame: DataFrame,
    waitTypeFieldName: string,
    newColumnName: string
): DataFrame {
    // Find the wait type field
    const waitTypeField = frame.fields.find(f => f.name === waitTypeFieldName);

    if (!waitTypeField) {
        console.warn(`Field ${waitTypeFieldName} not found`);
        return frame;
    }

    // Create new HTML field
    const htmlValues: string[] = waitTypeField.values.map(value => {
        let waitTypes: WaitTypeData[] = [];

        // Parse the wait type data (adjust based on your format)
        if (typeof value === 'string') {
            try {
                const parsed = JSON.parse(value);
                waitTypes = Object.entries(parsed).map(([type, count]) => ({
                    type,
                    count: count as number
                }));
            } catch (e) {
                return '';
            }
        } else if (typeof value === 'object' && value !== null) {
            if (Array.isArray(value)) {
                waitTypes = value;
            } else {
                waitTypes = Object.entries(value).map(([type, count]) => ({
                    type,
                    count: count as number
                }));
            }
        }

        return generateWaitTypeHTML(waitTypes);
    });

    const htmlField: Field = {
        name: newColumnName,
        type: FieldType.string,
        config: {},
        values: htmlValues,
    };

    // Add the new field to the frame
    return {
        ...frame,
        fields: [...frame.fields, htmlField]
    };
}
