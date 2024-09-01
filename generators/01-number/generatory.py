def numbers_gen(n):
    num = 1
    while num <= n:
        print(f'yielding number: {num}')
        yield num
        num += 1


def main():
    for num in numbers_gen(20):
        print(f'number received: {num}')
        print()


if __name__ == '__main__':
    main()
