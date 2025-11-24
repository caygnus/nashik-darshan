import 'package:test/test.dart';
import 'package:openapi/openapi.dart';


/// tests for AuthApi
void main() {
  final instance = Openapi().getAuthApi();

  group(AuthApi, () {
    // Signup
    //
    // Signup
    //
    //Future<DtoSignupResponse> authSignupPost(DtoSignupRequest signupRequest) async
    test('test authSignupPost', () async {
      // TODO
    });

  });
}
