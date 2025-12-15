import React, {useEffect, useState} from "react";
import {ExecutionPlanViewer, ParsedExecutionPlan} from "../ExecutionPlanTree/plan";
import {Card} from "@grafana/ui";
import {lastValueFrom, Observable} from "rxjs";
import {FetchResponse, getBackendSrv} from "@grafana/runtime";


export const QueryDetails: React.FC<{ sampleID: string, snapID: string }> = ({sampleID, snapID}) => {

    const [executionPlan, setExecutionPlan] = useState<ParsedExecutionPlan | null>(null);
    const [sample] = useState<string>(sampleID);
    const [snap] = useState<string>(snapID);

    // Function to fetch HTML content from backend
    useEffect(() => {
        const fetchPlan = async () => {
            if (!sample) {
                return;
            }
            try {
                let response: Observable<FetchResponse<ParsedExecutionPlan>>;
                response = await getBackendSrv().fetch({
                    url: '/api/plugins/guilhermearpassos-sqlsights-app/resources/getExecPlan?sampleId=' + sample + '&snapId=' + snap,
                });
                // Get the response as text since it's HTML
                const textResponse: { data: ParsedExecutionPlan } = await lastValueFrom(response);
                setExecutionPlan(textResponse.data);
            } catch (err) {
                throw new Error(`Failed to fetch: ${err}`);
            }
        };
        fetchPlan()
    }, [sample, snap]);
    return (<Card>
        <Card.Heading>Sample Details</Card.Heading>
        <Card.Description>
            {executionPlan && (<ExecutionPlanViewer executionPlan={executionPlan}/>)}
        </Card.Description>
    </Card>)
}
