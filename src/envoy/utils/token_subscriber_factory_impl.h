// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#pragma once

#include "src/envoy/utils/token_subscriber_factory.h"

namespace Envoy {
namespace Extensions {
namespace Utils {

class TokenSubscriberFactoryImpl : public TokenSubscriberFactory {
 public:
  TokenSubscriberFactoryImpl(Server::Configuration::FactoryContext& context)
      : context_(context) {}

  TokenSubscriberPtr createTokenSubscriber(
      const std::string& token_cluster, const std::string& token_url,
      const bool json_response,
      TokenSubscriber::TokenUpdateFunc callback) const override {
    return std::make_unique<TokenSubscriber>(context_, token_cluster, token_url,
                                             json_response, callback);
  }

  IamTokenSubscriberPtr createIamTokenSubscriber(
      IamTokenSubscriber::TokenGetFunc access_token_fn,
      const std::string& iam_service_cluster,
      const std::string& iam_service_uri,
      IamTokenSubscriber::TokenUpdateFunc callback) const override {
    return std::make_unique<IamTokenSubscriber>(context_, access_token_fn,
                                                iam_service_cluster,
                                                iam_service_uri, callback);
  }

 private:
  Envoy::Server::Configuration::FactoryContext& context_;
};

}  // namespace Utils
}  // namespace Extensions
}  // namespace Envoy
