import copy
class Super:
    def __init__(self):
        self.val = 0
        pass

    def copy(self): return copy.deepcopy(self)

class Test(Super):
    def __init__(self):
        self.value = 1
        super().__init__()
        pass
    def __del__(self): print("deleting")

t = Test()
print(t.value)
def test(c):
    print(c.value)
    del c
    return
test(t)
print(t.value)
t2 = copy.copy(t)
t.value = 3
print(t.value, t2.value)

t3 = t.copy()
t3.value = 10
t3.val = 10
print(t.value, t3.value, t.val, t3.val)
