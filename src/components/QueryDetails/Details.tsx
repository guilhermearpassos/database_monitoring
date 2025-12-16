import React, {useEffect, useState} from "react";
import {ExecutionPlan, ExecutionPlanViewer, ParsedExecutionPlan} from "../ExecutionPlanTree/plan";
import {Card} from "@grafana/ui";
import {lastValueFrom, Observable} from "rxjs";
import {FetchResponse, getBackendSrv} from "@grafana/runtime";
import {BlockingChain, BlockingChainComponent} from "./blockingChain";

interface QueryDetails {
    plan?: ParsedExecutionPlan;
    blocking_chain?: BlockingChain
}


export const QueryDetails: React.FC<{ sampleID: string, snapID: string }> = ({sampleID, snapID}) => {

    const [queryDetails, setQueryDetails] = useState<QueryDetails | null>(null);
    const [sample] = useState<string>(sampleID);
    const [snap] = useState<string>(snapID);

    // Function to fetch HTML content from backend
    useEffect(() => {
        const fetchPlan = async () => {
            if (!sample) {
                return;
            }
            try {
                let response: Observable<FetchResponse<QueryDetails>>;
                response = await getBackendSrv().fetch({
                    url: '/api/plugins/guilhermearpassos-sqlsights-app/resources/getQueryDetails?sampleId=' + sample + '&snapId=' + snap,
                });
                // Get the response as text since it's HTML
                const textResponse: { data: QueryDetails } = await lastValueFrom(response);
                setQueryDetails(textResponse.data);
            } catch (err) {
                throw new Error(`Failed to fetch: ${err}`);
            }
        };
        fetchPlan()
    }, [sample, snap]);
    return (<Card>
        <Card.Heading>Sample Details</Card.Heading>
        <Card.Description>
            {queryDetails && queryDetails.blocking_chain && queryDetails.blocking_chain.roots.length > 0 && (<BlockingChainComponent chain={queryDetails.blocking_chain} currentSampleId={sample}/>)}
            {queryDetails && queryDetails.plan && (<ExecutionPlanViewer executionPlan={queryDetails.plan}/>)}
        </Card.Description>
    </Card>)
}
