interface Hander {
    Success: (data:any) => void,
    Failture: (error:any) => void,
}

interface MetricQueryRequest extends Hander {
    Start: number,
    End: number,
    Query:string,
}

interface OfflineMachineRequest extends Hander {
    Cluster: string
    Hosts: string[]
}

interface NodeTopoRequest extends Hander {
    Cluster: string
    Start: number,
    End: number,
}

const requestMode:any = window.location.hostname=="localhost"?'cors':undefined;

class MonitorApi {
    targetURL:string = window.location.protocol + '//' + window.location.hostname + ':7096'; // 'http://monitor.default.svc.cluster.local:7096';
    baseURL:string = window.location.protocol + '//' + window.location.hostname + ':7098/api/admin/proxy';  // 'http://monitor.default.svc.cluster.local:7098';
    getHeaders = () => {
        return {
            'Content-Type': 'application/json',
            'X-Proxy-Target': this.targetURL,
        }
    }

    QueryMetric = (req:MetricQueryRequest) => {
        if (!req || !req.Query) {
            return
        }
        console.log("query:", req.Query);
        fetch(this.baseURL + '/query?start=' + req.Start + '&end=' + req.End + '&q=' + encodeURIComponent(req.Query), {
            method: 'GET',
            mode: requestMode,
            headers: this.getHeaders(),
        })
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }

    OfflineMachine = (req:OfflineMachineRequest) => {
        fetch(this.baseURL + '/api/resources/hosts/actions/offline', {
            method: 'POST',
            headers: this.getHeaders(),
            mode: requestMode,
            body: JSON.stringify({
                clusterName: req.Cluster,
                hostIPs: req.Hosts,
            })
        })
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }

    NodeTopo = (req:NodeTopoRequest) => {
        if (!req || !req.Cluster) {
            return
        }
        fetch(this.baseURL + '/api/node/topology?start=' + req.Start + '&end=' + req.End + '&clusterName=' + encodeURIComponent(req.Cluster), {
            method: 'GET',
            mode: requestMode,
            headers: this.getHeaders(),
        })
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }
}

interface KafkaPushRequest extends Hander {
    Topics:string,
    Payload:string,
}

interface ESIndicesRequest extends Hander {
    Wildcard?:string,
}

interface EnvsRequest extends Hander {
    Key?:string,
}

class AdminApi {
    baseURL:string = window.location.protocol + '//' + window.location.hostname + ':7098'; // 'http://monitor.default.svc.cluster.local:7098';
    getHeaders = () => {
        return {
            'Content-Type': 'application/json',
        }
    }

    KafkaPush =  (req:KafkaPushRequest) => {
        fetch(this.baseURL + '/api/admin/kafka/push', {
            method: 'POST',
            headers: this.getHeaders(),
            mode: requestMode,
            body: JSON.stringify({
                topics: req.Topics,
                payload: req.Payload,
            })
        })
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }
    Version =  (req:Hander) => {
        fetch(this.baseURL + '/api/admin/version')
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }
    ESIndices = (req:ESIndicesRequest) => {
        let path = '/api/admin/es/indices';
        if(!!req.Wildcard) {
            path += '/' + req.Wildcard;
        }
        fetch(this.baseURL + path + '?format=json', {
            method: 'GET',
            headers: this.getHeaders(),
            mode: requestMode,
        })
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }
    ESDeleteIndices = (req:ESIndicesRequest) => {
        let path = '/api/admin/es/indices/' + req.Wildcard;
        fetch(this.baseURL + path, {
            method: 'DELETE',
            headers: this.getHeaders(),
            mode: requestMode,
        })
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }
    Envs = (req:EnvsRequest) => {
        let path = '/api/admin/envs?format=json';
        if (!!req.Key) {
            path += "&key=" + encodeURIComponent(req.Key);
        }
        fetch(this.baseURL + path, {
            method: 'GET',
            headers: this.getHeaders(),
            mode: requestMode,
        })
        .then(res => res.json())
        .then(req.Success)
        .catch(req.Failture);
    }
}

export { MonitorApi, AdminApi };
