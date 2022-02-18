/*
 * Copyright (c) 2021 Terminus, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


UPDATE `pipeline_configs`
SET `value` ='{\n \"kind\": \"SCHEDULER\",\n \"name\": \"scheduler\",\n \"options\": {\n  \"ADDR\": \"orchestrator.marathon.l4lb.thisdcos.directory:8081\"\n }\n}'
WHERE `id` = 1 AND `type` = 'action_executor';