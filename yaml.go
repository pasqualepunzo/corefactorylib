package lib

import (
	"encoding/json"
	"strconv"
	"strings"
)

func GetYamlContainerProbes(prbsarray []Probes) string {

	yamlProbes := ""

	if prbsarray == nil {
		return ""
	}

	for _, prbs := range prbsarray {
		switch prbs.Category {
		case "startup":
			yamlProbes += "        startupProbe:\n"
			break
		case "liveness":
			yamlProbes += "        livenessProbe:\n"
			break
		case "readiness":
			yamlProbes += "        readinessProbe:\n"
			break

			// TODO: gestire default: mettere log di errore
		}

		yamlProbes += "          " + prbs.Type + ":\n"
		if prbs.Type == "exec" {
			yamlProbes += "            command:\n"
			var cmdarray []string
			err := json.Unmarshal([]byte(prbs.Command), &cmdarray)
			//Logga(ctx, httpheadersarray)
			//Logga(ctx, err)
			if err == nil {
				for _, cmdvalue := range cmdarray {
					if cmdvalue != "" {
						//yamlProbes += "            - " + cmdvalue + "\n"
						yamlProbes += "            - " + strings.Replace(cmdvalue, "_DQ_", "\"", -1) + "\n"
					}
				}
			}
		} else if prbs.Type == "httpGet" {
			if prbs.HttpHost != "" {
				yamlProbes += "            host: " + prbs.HttpHost + "\n"
			}
			yamlProbes += "            path: " + prbs.HttpPath + "\n"
			yamlProbes += "            port: " + strconv.Itoa(prbs.HttpPort) + "\n"
			yamlProbes += "            scheme: " + prbs.HttpScheme + "\n"
			if prbs.HttpHeaders != "" {
				yamlProbes += "            httpHeaders:\n"

				var httpheadersarray []HttpHeadersJson
				err := json.Unmarshal([]byte(prbs.HttpHeaders), &httpheadersarray)
				//Logga(ctx, httpheadersarray)
				//Logga(ctx, err)
				if err == nil {
					for _, HttpNameValue := range httpheadersarray {
						if HttpNameValue.Name != "" {
							yamlProbes += "            - name: " + HttpNameValue.Name + "\n"
							yamlProbes += "              value: " + HttpNameValue.Value + "\n"
						}
					}
				}
			}
		} else if prbs.Type == "tcpSocket" {
			if prbs.TcpHost != "" {
				yamlProbes += "            host: " + prbs.TcpHost + "\n"
			}
			yamlProbes += "            port: " + strconv.Itoa(prbs.TcpPort) + "\n"
		} else if prbs.Type == "grpc" {
			yamlProbes += "            port: " + strconv.Itoa(prbs.GrpcPort) + "\n"
		}

		yamlProbes += "          initialDelaySeconds: " + strconv.Itoa(prbs.InitialDelaySeconds) + "\n"
		yamlProbes += "          periodSeconds: " + strconv.Itoa(prbs.PeriodSeconds) + "\n"
		yamlProbes += "          timeoutSeconds: " + strconv.Itoa(prbs.TimeoutSeconds) + "\n"
		yamlProbes += "          successThreshold: " + strconv.Itoa(prbs.SuccessThreshold) + "\n"
		yamlProbes += "          failureThreshold: " + strconv.Itoa(prbs.FailureThreshold) + "\n"

		// fmt.Println(yamlProbes)
		// os.Exit(0)
	}

	return yamlProbes
}
func GetYamlHpa(ires IstanzaMicro, micros Microservice, versione string) string {

	versioneApp := versione

	namespace := ""
	if ires.Microservice == "msrefappmonolith" {
		if ires.SwMultiEnvironment == "1" {
			namespace = micros.Namespace
		} else {
			namespace = "monolith"
		}
	} else {
		namespace = micros.Namespace
	}

	// oggi porcata domani sarebbe opportuno aplicare un flag in kubemicroserv
	switch micros.Nome {
	case "mscorequery", "mscorewrite", "mscoreservice":
		versioneApp = "latest"
		break
	}

	app := ""
	if ires.Microservice == "msrefappmonolith" {
		app = ires.PodName + "-v" + versione
	} else {
		app = micros.Nome + "-v" + versioneApp
	}

	yamlHpa := "apiVersion: autoscaling/v2beta2\n"
	yamlHpa += "kind: HorizontalPodAutoscaler\n"
	yamlHpa += "metadata:\n"
	yamlHpa += "  name: " + app + "\n"
	yamlHpa += "  namespace: " + namespace + "\n"
	yamlHpa += "spec:\n"
	yamlHpa += "  scaleTargetRef:\n"
	yamlHpa += "    apiVersion: apps/v1\n"
	yamlHpa += "    kind: Deployment\n"
	yamlHpa += "    name: " + app + "\n"
	yamlHpa += "  minReplicas: " + micros.Hpa.MinReplicas + "\n"
	yamlHpa += "  maxReplicas: " + micros.Hpa.MaxReplicas + "\n"

	yamlHpa += "  metrics:\n"
	yamlHpa += "  - type: Resource\n"
	yamlHpa += "    resource:\n"
	yamlHpa += "      name: cpu\n"
	yamlHpa += "      target:\n"
	yamlHpa += "        type: Utilization\n"
	yamlHpa += "        averageUtilization: " + micros.Hpa.CpuTarget + "\n"
	yamlHpa += "  - type: Resource\n"
	yamlHpa += "    resource:\n"
	yamlHpa += "      name: memory \n"
	yamlHpa += "      target:\n"
	yamlHpa += "        type: Utilization\n"
	yamlHpa += "        averageUtilization: " + micros.Hpa.MemTarget + "\n"
	yamlHpa += "\n---\n\n"

	return yamlHpa
}
