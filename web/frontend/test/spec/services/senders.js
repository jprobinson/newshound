'use strict';

describe('Service: senderColors', function () {

  // load the service's module
  beforeEach(module('newshoundApp'));

  // instantiate service
  var senderColors;
  beforeEach(inject(function (_senderColors_) {
    senderColors = _senderColors_;
  }));

  it('should do something', function () {
    expect(!!senderColors).toBe(true);
  });

});
