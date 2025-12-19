import {DataFrame, Field, FieldType} from '@grafana/data';
import {getWaitEventColor, WaitTypeData} from "../utils/wait_event_colors";


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
